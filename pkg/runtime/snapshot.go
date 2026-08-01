package runtime

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"sort"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// Datastore snapshot file format, version 1. The layout below is the format
// definition.
//
//	magic       [8]  "DUSOSNAP"
//	version     u16
//	flags       u16   bit0 = body encrypted
//	reserved    u32
//	seq         u64   WAL watermark this snapshot covers
//	--- body, encrypted as a unit when flags bit0 is set ---
//	keycount    u32   + N × (varint keylen + key + codec value)
//	expirycount u32   + N × (varint keylen + key + i64 unix millis)
//
// The header stays plaintext so version, flags and the WAL watermark stay
// readable without the encryption key — a replica needs the watermark to know
// where to resume, and a corrupt-file diagnosis shouldn't require the key.
const (
	snapshotMagic         = "DUSOSNAP"
	snapshotVersion       = 1
	snapshotFlagEncrypted = 1 << 0

	// magic(8) + version(2) + flags(2) + reserved(4) + seq(8)
	snapshotHeaderLen = 24
)

// marshalSnapshotBody encodes the key/value and expiry sections. Keys are
// written in sorted order so identical state produces identical bytes, which
// makes snapshots comparable across nodes.
func marshalSnapshotBody(data map[string]any, expiry map[string]time.Time) ([]byte, error) {
	buf := make([]byte, 0, 64*len(data)+32)

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(keys)))
	var err error
	for _, k := range keys {
		buf = binary.AppendUvarint(buf, uint64(len(k)))
		buf = append(buf, k...)
		if buf, err = script.EncodeValue(buf, data[k]); err != nil {
			return nil, fmt.Errorf("key %q: %v", k, err)
		}
	}

	ekeys := make([]string, 0, len(expiry))
	for k := range expiry {
		ekeys = append(ekeys, k)
	}
	sort.Strings(ekeys)

	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(ekeys)))
	for _, k := range ekeys {
		buf = binary.AppendUvarint(buf, uint64(len(k)))
		buf = append(buf, k...)
		buf = binary.LittleEndian.AppendUint64(buf, uint64(expiry[k].UnixMilli()))
	}

	return buf, nil
}

func unmarshalSnapshotBody(body []byte) (map[string]any, map[string]time.Time, error) {
	if len(body) < 4 {
		return nil, nil, fmt.Errorf("truncated snapshot: missing key count")
	}
	count := binary.LittleEndian.Uint32(body)
	body = body[4:]

	data := make(map[string]any, count)
	for i := uint32(0); i < count; i++ {
		key, rest, err := readSnapshotKey(body)
		if err != nil {
			return nil, nil, fmt.Errorf("snapshot entry %d: %v", i, err)
		}
		value, rest, err := script.DecodeValue(rest)
		if err != nil {
			return nil, nil, fmt.Errorf("snapshot key %q: %v", key, err)
		}
		data[key] = value
		body = rest
	}

	// The expiry section is mandatory in v1, so a short read here is truncation,
	// not an older layout. A format that drops or moves the section would carry a
	// different version, which decodeSnapshot already rejects.
	if len(body) < 4 {
		return nil, nil, fmt.Errorf("truncated snapshot: missing expiry count")
	}
	ecount := binary.LittleEndian.Uint32(body)
	body = body[4:]

	expiry := make(map[string]time.Time, ecount)
	for i := uint32(0); i < ecount; i++ {
		key, rest, err := readSnapshotKey(body)
		if err != nil {
			return nil, nil, fmt.Errorf("snapshot expiry entry %d: %v", i, err)
		}
		if len(rest) < 8 {
			return nil, nil, fmt.Errorf("snapshot expiry for key %q: truncated deadline", key)
		}
		expiry[key] = time.UnixMilli(int64(binary.LittleEndian.Uint64(rest)))
		body = rest[8:]
	}

	return data, expiry, nil
}

func readSnapshotKey(buf []byte) (string, []byte, error) {
	n, adv := binary.Uvarint(buf)
	if adv <= 0 {
		return "", nil, fmt.Errorf("malformed key length")
	}
	buf = buf[adv:]
	if n > uint64(len(buf)) {
		return "", nil, fmt.Errorf("key length %d exceeds %d remaining bytes", n, len(buf))
	}
	return string(buf[:n]), buf[n:], nil
}

// encodeSnapshot builds a complete v1 snapshot file for the store's current
// state. Callers must hold dataMutex.
func (ds *DatastoreValue) encodeSnapshot() ([]byte, error) {
	body, err := marshalSnapshotBody(ds.data, ds.expiryTimes)
	if err != nil {
		return nil, err
	}

	var flags uint16
	if len(ds.encryptKey) > 0 {
		flags |= snapshotFlagEncrypted
		if body, err = ds.encryptBytes(body); err != nil {
			return nil, err
		}
	}

	out := make([]byte, 0, snapshotHeaderLen+len(body))
	out = append(out, snapshotMagic...)
	out = binary.LittleEndian.AppendUint16(out, snapshotVersion)
	out = binary.LittleEndian.AppendUint16(out, flags)
	out = binary.LittleEndian.AppendUint32(out, 0) // reserved
	out = binary.LittleEndian.AppendUint64(out, ds.walSeq.Load())
	return append(out, body...), nil
}

// decodeSnapshot parses a snapshot file, transparently handling files written
// by pre-v1 duso (a bare gob map with no header at all).
func (ds *DatastoreValue) decodeSnapshot(raw []byte) (map[string]any, map[string]time.Time, uint64, error) {
	if !bytes.HasPrefix(raw, []byte(snapshotMagic)) {
		data, err := ds.decodeLegacySnapshot(raw)
		if err != nil {
			return nil, nil, 0, err
		}
		// Pre-v1 files carry no deadlines and no watermark.
		return data, make(map[string]time.Time), 0, nil
	}

	if len(raw) < snapshotHeaderLen {
		return nil, nil, 0, fmt.Errorf("truncated snapshot header")
	}
	version := binary.LittleEndian.Uint16(raw[8:10])
	if version != snapshotVersion {
		return nil, nil, 0, fmt.Errorf("snapshot %q was written by a newer duso (format version %d, this build understands %d)",
			ds.persistPath, version, snapshotVersion)
	}
	flags := binary.LittleEndian.Uint16(raw[10:12])
	seq := binary.LittleEndian.Uint64(raw[16:24])
	body := raw[snapshotHeaderLen:]

	if flags&snapshotFlagEncrypted != 0 {
		if len(ds.encryptKey) == 0 {
			return nil, nil, 0, fmt.Errorf("snapshot %q is encrypted but no encrypt_key is configured", ds.persistPath)
		}
		var err error
		if body, err = ds.decryptBytes(body); err != nil {
			return nil, nil, 0, err
		}
	} else if len(ds.encryptKey) > 0 {
		return nil, nil, 0, fmt.Errorf("snapshot %q is not encrypted but an encrypt_key is configured", ds.persistPath)
	}

	data, expiry, err := unmarshalSnapshotBody(body)
	if err != nil {
		return nil, nil, 0, err
	}
	return data, expiry, seq, nil
}

// decodeLegacySnapshot reads a pre-v1 file: a bare gob-encoded map, optionally
// encrypted as a whole. Deprecated — remove once the migration window closes.
func (ds *DatastoreValue) decodeLegacySnapshot(raw []byte) (map[string]any, error) {
	gobBytes, err := ds.decryptBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt legacy datastore %q: %v", ds.namespace, err)
	}

	data := make(map[string]any)
	if err := gob.NewDecoder(bytes.NewReader(gobBytes)).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to deserialize legacy datastore %q: %v", ds.namespace, err)
	}
	return data, nil
}

// applySnapshotExpiry merges deadlines from a decoded snapshot into the live
// expiry state. Callers must hold dataMutex.
//
// Deadlines already in the past are kept rather than dropped: the sweeper
// removes them on its next tick, and that tick is what produces the
// authoritative deletion record in the WAL. Dropping them here would delete the
// key without ever logging why.
func (ds *DatastoreValue) applySnapshotExpiry(expiry map[string]time.Time) {
	for k, deadline := range expiry {
		if _, exists := ds.data[k]; !exists {
			continue // deadline for a key the snapshot no longer holds
		}
		ds.expiryTimes[k] = deadline
		// Stale heap entries for a re-expired key are skipped by the sweeper's
		// lazy check, so pushing unconditionally is safe.
		heap.Push(&ds.expiryHeap, ExpiryEntry{key: k, expiryTime: deadline})
	}
}
