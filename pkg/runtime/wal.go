package runtime

import (
	"bufio"
	"container/heap"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// Datastore write-ahead log, format version 1. The layout and opcode table
// below are the format definition.
//
//	file header (16 bytes, written once)
//	  magic     [8]  "DUSOWAL\0"
//	  version   u16
//	  flags     u16   bit0 = record bodies encrypted
//	  reserved  u32
//
//	record frame
//	  length    u32   body byte count
//	  crc32c    u32   over the body bytes as stored
//	  body      [length]
//
//	record body
//	  seq       varint
//	  op        u8
//	  ...then exactly the fields the opcode's shape specifies
//
// This is an operation log, not a state log: SHIFT records the fact of a shift,
// not the resulting array. That is what keeps queue operations O(1) on disk, and
// it is what makes the log replayable as a deterministic stream of operations.
const (
	walMagic         = "DUSOWAL\x00"
	walVersion       = 1
	walFlagEncrypted = 1 << 0

	// magic(8) + version(2) + flags(2) + reserved(4)
	walHeaderLen = 16

	// Frame prefix: length(4) + crc32c(4)
	walFrameLen = 8
)

// walOp identifies an operation. Codes are allocated in reserved ranges so an
// unrecognized opcode can still be classified: below 0x80 is data (an unknown
// one is fatal, since skipping it would silently diverge state), at or above
// 0x80 is control/metadata (an unknown one is safe to skip).
type walOp uint8

const (
	// key-value ops, 0x01-0x2F
	opSet     walOp = 0x01
	opSetOnce walOp = 0x02
	opDelete  walOp = 0x03
	opSwap    walOp = 0x04
	opUpdate  walOp = 0x05
	opIncr    walOp = 0x06
	opRename  walOp = 0x07
	opExpire  walOp = 0x08
	opExpired walOp = 0x09

	// array / queue ops, 0x30-0x4F
	opPush    walOp = 0x30
	opUnshift walOp = 0x31
	opShift   walOp = 0x32
	opPop     walOp = 0x33

	// store-wide ops, 0x50-0x5F
	opClear walOp = 0x50

	// walOpDataLimit is the boundary between data and control opcodes.
	walOpDataLimit walOp = 0x80
)

// walShape is the normative record layout for an opcode. There is no
// field-presence bitmask in the record: the opcode determines the shape, so a
// record cannot contradict itself.
type walShape struct {
	name    string
	hasKey  bool
	hasKey2 bool
	hasNum  bool
	hasVal  bool
}

var walShapes = map[walOp]walShape{
	opSet:     {name: "SET", hasKey: true, hasVal: true},
	opSetOnce: {name: "SET_ONCE", hasKey: true, hasVal: true},
	opDelete:  {name: "DELETE", hasKey: true},
	opSwap:    {name: "SWAP", hasKey: true, hasVal: true},
	opUpdate:  {name: "UPDATE", hasKey: true, hasVal: true},
	opIncr:    {name: "INCR", hasKey: true, hasNum: true},
	opRename:  {name: "RENAME", hasKey: true, hasKey2: true},
	opExpire:  {name: "EXPIRE", hasKey: true, hasNum: true},
	opExpired: {name: "EXPIRED", hasKey: true},
	opPush:    {name: "PUSH", hasKey: true, hasVal: true},
	opUnshift: {name: "UNSHIFT", hasKey: true, hasVal: true},
	opShift:   {name: "SHIFT", hasKey: true},
	opPop:     {name: "POP", hasKey: true},
	opClear:   {name: "CLEAR"},
}

// crc32c catches torn and corrupted records. It is computed over the body
// exactly as stored, so an encrypted record can be integrity-checked without
// the key.
var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

// errSkippableRecord marks an unrecognized control record (opcode >= 0x80). The
// frame length lets the reader step over it without understanding it.
var errSkippableRecord = errors.New("skippable control record")

type walRecord struct {
	Seq   uint64
	Op    walOp
	Key   string
	Key2  string
	Num   float64
	Value any
}

func appendWALString(buf []byte, s string) []byte {
	buf = binary.AppendUvarint(buf, uint64(len(s)))
	return append(buf, s...)
}

func readWALString(buf []byte) (string, []byte, error) {
	n, adv := binary.Uvarint(buf)
	if adv <= 0 {
		return "", nil, fmt.Errorf("malformed string length")
	}
	buf = buf[adv:]
	if n > uint64(len(buf)) {
		return "", nil, fmt.Errorf("string length %d exceeds %d remaining bytes", n, len(buf))
	}
	return string(buf[:n]), buf[n:], nil
}

func encodeWALBody(rec *walRecord) ([]byte, error) {
	shape, ok := walShapes[rec.Op]
	if !ok {
		return nil, fmt.Errorf("refusing to write unknown WAL opcode 0x%02X", byte(rec.Op))
	}

	buf := binary.AppendUvarint(make([]byte, 0, 32), rec.Seq)
	buf = append(buf, byte(rec.Op))

	if shape.hasKey {
		buf = appendWALString(buf, rec.Key)
	}
	if shape.hasKey2 {
		buf = appendWALString(buf, rec.Key2)
	}
	if shape.hasNum {
		buf = binary.LittleEndian.AppendUint64(buf, math.Float64bits(rec.Num))
	}
	if shape.hasVal {
		var err error
		if buf, err = script.EncodeValue(buf, rec.Value); err != nil {
			return nil, fmt.Errorf("%s record for key %q: %v", shape.name, rec.Key, err)
		}
	}
	return buf, nil
}

func decodeWALBody(body []byte) (*walRecord, error) {
	seq, adv := binary.Uvarint(body)
	if adv <= 0 {
		return nil, fmt.Errorf("malformed record sequence number")
	}
	body = body[adv:]
	if len(body) == 0 {
		return nil, fmt.Errorf("record has no opcode")
	}

	rec := &walRecord{Seq: seq, Op: walOp(body[0])}
	body = body[1:]

	shape, known := walShapes[rec.Op]
	if !known {
		if rec.Op >= walOpDataLimit {
			// Control or vendor record from a newer or cluster-aware writer.
			// Ignoring it cannot affect key-value state.
			return rec, errSkippableRecord
		}
		return nil, fmt.Errorf("unknown WAL data opcode 0x%02X — this file was written by a newer duso", byte(rec.Op))
	}

	var err error
	if shape.hasKey {
		if rec.Key, body, err = readWALString(body); err != nil {
			return nil, fmt.Errorf("%s record: key: %v", shape.name, err)
		}
	}
	if shape.hasKey2 {
		if rec.Key2, body, err = readWALString(body); err != nil {
			return nil, fmt.Errorf("%s record: second key: %v", shape.name, err)
		}
	}
	if shape.hasNum {
		if len(body) < 8 {
			return nil, fmt.Errorf("%s record: truncated number", shape.name)
		}
		rec.Num = math.Float64frombits(binary.LittleEndian.Uint64(body))
		body = body[8:]
	}
	if shape.hasVal {
		if rec.Value, body, err = script.DecodeValue(body); err != nil {
			return nil, fmt.Errorf("%s record for key %q: %v", shape.name, rec.Key, err)
		}
	}
	if len(body) != 0 {
		return nil, fmt.Errorf("%s record for key %q has %d unexpected trailing bytes", shape.name, rec.Key, len(body))
	}
	return rec, nil
}

// writeWALOp appends one operation to the log. Callers hold dataMutex, so log
// order matches the order operations are applied to memory — without that the
// log could describe a different final state than the store holds.
func (ds *DatastoreValue) writeWALOp(op walOp, key, key2 string, num float64, value any) error {
	if ds.walPath == "" || ds.walFile == nil {
		return nil // WAL not configured
	}

	ds.walMutex.Lock()
	defer ds.walMutex.Unlock()

	rec := &walRecord{Seq: ds.walSeq.Add(1), Op: op, Key: key, Key2: key2, Num: num, Value: value}
	body, err := encodeWALBody(rec)
	if err != nil {
		return err
	}
	if len(ds.encryptKey) > 0 {
		if body, err = ds.encryptBytes(body); err != nil {
			return fmt.Errorf("failed to encrypt WAL record for key %q: %v", key, err)
		}
	}

	frame := make([]byte, walFrameLen, walFrameLen+len(body))
	binary.LittleEndian.PutUint32(frame[0:4], uint32(len(body)))
	binary.LittleEndian.PutUint32(frame[4:8], crc32.Checksum(body, crc32cTable))
	frame = append(frame, body...)

	if _, err := ds.walFile.Write(frame); err != nil {
		return fmt.Errorf("failed to write WAL record for key %q: %v", key, err)
	}
	if ds.walSyncInterval == 0 {
		if err := ds.walFile.Sync(); err != nil {
			return fmt.Errorf("failed to sync WAL: %v", err)
		}
	}
	return nil
}

// appendWALHeader writes the 16-byte file header. Called when the log is created
// or truncated.
func (ds *DatastoreValue) appendWALHeader(f *os.File) error {
	var flags uint16
	if len(ds.encryptKey) > 0 {
		flags |= walFlagEncrypted
	}
	hdr := make([]byte, 0, walHeaderLen)
	hdr = append(hdr, walMagic...)
	hdr = binary.LittleEndian.AppendUint16(hdr, walVersion)
	hdr = binary.LittleEndian.AppendUint16(hdr, flags)
	hdr = binary.LittleEndian.AppendUint32(hdr, 0) // reserved

	if _, err := f.Write(hdr); err != nil {
		return fmt.Errorf("failed to write WAL header: %v", err)
	}
	return f.Sync()
}

// readWALHeader validates an existing header and reports whether bodies are
// encrypted.
func (ds *DatastoreValue) readWALHeader(hdr []byte) (encrypted bool, err error) {
	if len(hdr) < walHeaderLen || string(hdr[:8]) != walMagic {
		return false, fmt.Errorf("not a duso WAL file")
	}
	version := binary.LittleEndian.Uint16(hdr[8:10])
	if version != walVersion {
		return false, fmt.Errorf("WAL %q was written by a newer duso (format version %d, this build understands %d)",
			ds.walPath, version, walVersion)
	}
	flags := binary.LittleEndian.Uint16(hdr[10:12])
	encrypted = flags&walFlagEncrypted != 0

	if encrypted && len(ds.encryptKey) == 0 {
		return false, fmt.Errorf("WAL %q is encrypted but no encrypt_key is configured", ds.walPath)
	}
	if !encrypted && len(ds.encryptKey) > 0 {
		return false, fmt.Errorf("WAL %q is not encrypted but an encrypt_key is configured", ds.walPath)
	}
	return encrypted, nil
}

// walReplayResult reports what a replay pass consumed.
type walReplayResult struct {
	applied    int   // records applied to memory
	lastGood   int64 // byte offset just past the last intact record
	tornAt     int64 // offset where an incomplete record began, -1 if the log ended cleanly
	tornReason string
}

// replayWALv1 streams the log and applies each record. Records at or below the
// snapshot watermark are skipped: the snapshot already contains them, and
// replaying a SHIFT or INCR twice would corrupt state.
//
// Reading is streamed rather than slurped so a log that has grown between
// snapshots cannot be a memory spike at startup.
//
// A crashed process almost always leaves a partial final record — it died
// mid-write, or mid-fsync. That is the normal ending for a write-ahead log, not
// an error: the interrupted write was never acknowledged as durable, so
// discarding it loses nothing a caller was promised. Replay therefore stops
// cleanly at a torn tail.
//
// Corruption is a different thing and stays fatal. The distinction the framing
// buys us:
//
//   - bytes are MISSING (short frame header, or a body that runs past the end of
//     the file) — the write never finished. Torn tail, stop cleanly.
//   - bytes are PRESENT but WRONG (checksum fails) and more records follow — the
//     damage is inside the log, not at its end, and everything after it is
//     unreachable because the framing can no longer be trusted. Fatal.
//   - checksum fails on the final record — all its bytes exist but disagree with
//     the checksum, which a half-flushed sector can produce. Torn tail.
//   - checksum passes but the body will not decode — the bytes are exactly what
//     the writer wrote, so this is a format problem, never tornness. Fatal.
func (ds *DatastoreValue) replayWALv1(f *os.File, encrypted bool, size int64) (walReplayResult, error) {
	watermark := ds.walSeq.Load()
	maxSeq := watermark
	res := walReplayResult{lastGood: walHeaderLen, tornAt: -1}

	r := bufio.NewReader(f)
	frame := make([]byte, walFrameLen)
	offset := int64(walHeaderLen)

	for {
		recordStart := offset

		if _, err := io.ReadFull(r, frame); err != nil {
			if err == io.EOF {
				break // clean end of log
			}
			// Fewer than 8 bytes left: a frame header that never finished.
			res.tornAt = recordStart
			res.tornReason = "incomplete record header"
			break
		}
		offset += walFrameLen

		length := int64(binary.LittleEndian.Uint32(frame[0:4]))
		wantCRC := binary.LittleEndian.Uint32(frame[4:8])
		if length > size-offset {
			res.tornAt = recordStart
			res.tornReason = fmt.Sprintf("record declares %d bytes but only %d remain", length, size-offset)
			break
		}

		body := make([]byte, length)
		if _, err := io.ReadFull(r, body); err != nil {
			res.tornAt = recordStart
			res.tornReason = "incomplete record body"
			break
		}
		offset += length
		isLastRecord := offset >= size

		if got := crc32.Checksum(body, crc32cTable); got != wantCRC {
			if isLastRecord {
				// All the bytes are there but disagree with the checksum, and
				// nothing follows: a half-flushed final write.
				res.tornAt = recordStart
				res.tornReason = fmt.Sprintf("checksum mismatch on the final record (want %08x, got %08x)", wantCRC, got)
				break
			}
			return res, fmt.Errorf("WAL record at offset %d failed its checksum (want %08x, got %08x) with %d bytes of log after it — the log is corrupt, not merely truncated",
				recordStart, wantCRC, got, size-offset)
		}

		if encrypted {
			plain, err := ds.decryptBytes(body)
			if err != nil {
				if isLastRecord {
					res.tornAt = recordStart
					res.tornReason = "final record failed to decrypt"
					break
				}
				return res, fmt.Errorf("failed to decrypt WAL record at offset %d: %v", recordStart, err)
			}
			body = plain
		}

		rec, err := decodeWALBody(body)
		if err != nil {
			if errors.Is(err, errSkippableRecord) {
				if rec.Seq > maxSeq {
					maxSeq = rec.Seq
				}
				res.lastGood = offset
				continue
			}
			// Checksum passed, so these are the writer's own bytes.
			return res, fmt.Errorf("WAL record at offset %d: %v", recordStart, err)
		}
		res.lastGood = offset

		if rec.Seq > maxSeq {
			maxSeq = rec.Seq
		}
		if rec.Seq <= watermark {
			continue // already captured by the snapshot
		}
		if err := ds.applyWALRecord(rec); err != nil {
			return res, fmt.Errorf("WAL record at offset %d: %v", recordStart, err)
		}
		res.applied++
	}

	ds.walSeq.Store(maxSeq)
	return res, nil
}

// applyWALRecord replays one operation directly against the in-memory maps.
// It deliberately bypasses the public mutation methods, which would log again.
// Callers hold dataMutex.
func (ds *DatastoreValue) applyWALRecord(rec *walRecord) error {
	switch rec.Op {
	case opSet, opSetOnce, opSwap:
		// SET_ONCE is only logged when it actually stored, so replay is a blind
		// write for all three.
		ds.data[rec.Key] = rec.Value

	case opDelete, opExpired:
		delete(ds.data, rec.Key)
		delete(ds.expiryTimes, rec.Key)

	case opClear:
		ds.data = make(map[string]any)
		ds.expiryTimes = make(map[string]time.Time)
		ds.expiryHeap = ds.expiryHeap[:0]

	case opUpdate:
		patch, ok := rec.Value.(map[string]any)
		if !ok {
			return fmt.Errorf("UPDATE record for key %q carries a non-object patch", rec.Key)
		}
		obj := make(map[string]any)
		if current, exists := ds.data[rec.Key]; exists {
			if o, ok := current.(map[string]any); ok {
				obj = DeepCopyAny(o).(map[string]any)
			}
		}
		deepMerge(obj, patch)
		ds.data[rec.Key] = obj

	case opIncr:
		current := 0.0
		if v, exists := ds.data[rec.Key]; exists {
			f, ok := v.(float64)
			if !ok {
				return fmt.Errorf("INCR record for key %q applies to a non-numeric value", rec.Key)
			}
			current = f
		}
		ds.data[rec.Key] = current + rec.Num

	case opRename:
		value, exists := ds.data[rec.Key]
		if !exists {
			return nil
		}
		ds.data[rec.Key2] = value
		delete(ds.data, rec.Key)
		if deadline, ok := ds.expiryTimes[rec.Key]; ok {
			ds.expiryTimes[rec.Key2] = deadline
			delete(ds.expiryTimes, rec.Key)
			heap.Push(&ds.expiryHeap, ExpiryEntry{key: rec.Key2, expiryTime: deadline})
		}

	case opExpire:
		if _, exists := ds.data[rec.Key]; !exists {
			return nil
		}
		// Absolute deadline, not a TTL: replaying hours later must not extend it.
		deadline := time.UnixMilli(int64(rec.Num))
		ds.expiryTimes[rec.Key] = deadline
		heap.Push(&ds.expiryHeap, ExpiryEntry{key: rec.Key, expiryTime: deadline})

	case opPush:
		arr, _ := ds.data[rec.Key].([]any)
		ds.data[rec.Key] = append(append(make([]any, 0, len(arr)+1), arr...), rec.Value)

	case opUnshift:
		arr, _ := ds.data[rec.Key].([]any)
		ds.data[rec.Key] = append(append(make([]any, 0, len(arr)+1), rec.Value), arr...)

	case opShift:
		if arr, ok := ds.data[rec.Key].([]any); ok && len(arr) > 0 {
			ds.data[rec.Key] = arr[1:]
		}

	case opPop:
		if arr, ok := ds.data[rec.Key].([]any); ok && len(arr) > 0 {
			ds.data[rec.Key] = arr[:len(arr)-1]
		}

	default:
		return fmt.Errorf("unhandled WAL opcode 0x%02X", byte(rec.Op))
	}
	return nil
}
