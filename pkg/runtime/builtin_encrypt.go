package runtime

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// EncryptedValue is a wrapper for gob encoding/decoding encrypted data
type EncryptedValue struct {
	Data any
}

func init() {
	gob.Register(EncryptedValue{})
	gob.Register(map[string]any{})
	gob.Register([]any{})
}

func init() {
	// Register types for gob encoding
	gob.Register(map[string]any{})
	gob.Register([]any{})
}

// builtinEncrypt encrypts data using AES-256-GCM
// Usage: encrypt(data, key)
// data: any duso value (string, number, array, object, etc.)
// key: 32-byte encryption key (string or binary)
// Returns: base64-encoded (nonce + ciphertext)
// The data is gob-encoded before encryption to preserve types and structure.
func builtinEncrypt(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get data (positional 0 or named "data")
	var dataArg any
	if d, ok := args["data"]; ok {
		dataArg = d
	} else if d, ok := args["0"]; ok {
		dataArg = d
	}
	if dataArg == nil {
		return nil, fmt.Errorf("encrypt() requires a data argument")
	}

	// Gob-encode the data to preserve types and structure
	dataBytes, err := gobEncode(dataArg)
	if err != nil {
		return nil, fmt.Errorf("encrypt() failed to encode data: %v", err)
	}
	if len(dataBytes) == 0 {
		return nil, fmt.Errorf("encrypt() requires non-empty data")
	}

	// Get key (positional 1 or named "key")
	var keyArg any
	if k, ok := args["key"]; ok {
		keyArg = k
	} else if k, ok := args["1"]; ok {
		keyArg = k
	}
	if keyArg == nil {
		return nil, fmt.Errorf("encrypt() requires a key argument")
	}

	// Try to decode as base64 first (standard format)
	keyStr := ""
	if str, ok := keyArg.(string); ok {
		keyStr = str
	} else {
		keyStr = fmt.Sprintf("%v", keyArg)
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("encrypt() key must be base64-encoded: %v", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encrypt() key must decode to 32 bytes for AES-256 (got %d)", len(keyBytes))
	}

	// Create AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("encrypt() failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encrypt() failed to create GCM: %v", err)
	}

	// Generate random 12-byte nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("encrypt() failed to generate nonce: %v", err)
	}

	// Encrypt: ciphertext includes the authentication tag
	ciphertext := gcm.Seal(nil, nonce, dataBytes, nil)

	// Prepend nonce to ciphertext for storage
	encryptedData := append(nonce, ciphertext...)

	// Return base64-encoded result
	encoded := base64.StdEncoding.EncodeToString(encryptedData)
	return encoded, nil
}

// builtinDecrypt decrypts data using AES-256-GCM
// Usage: decrypt(encrypted_data, key)
// encrypted_data: base64-encoded (nonce + ciphertext) from encrypt()
// key: 32-byte encryption key (same key used for encryption)
// Returns: decrypted value with original type preserved (string, array, object, etc.)
func builtinDecrypt(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get encrypted data (positional 0 or named "encrypted_data")
	var dataArg any
	if d, ok := args["encrypted_data"]; ok {
		dataArg = d
	} else if d, ok := args["0"]; ok {
		dataArg = d
	}
	if dataArg == nil {
		return nil, fmt.Errorf("decrypt() requires an encrypted_data argument")
	}

	// Extract base64 string
	var encodedStr string
	if str, ok := dataArg.(string); ok {
		encodedStr = str
	} else {
		encodedStr = fmt.Sprintf("%v", dataArg)
	}

	if encodedStr == "" {
		return nil, fmt.Errorf("decrypt() requires non-empty encrypted_data")
	}

	// Decode base64
	encryptedData, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		return nil, fmt.Errorf("decrypt() invalid base64: %v", err)
	}

	// Get key (positional 1 or named "key")
	var keyArg any
	if k, ok := args["key"]; ok {
		keyArg = k
	} else if k, ok := args["1"]; ok {
		keyArg = k
	}
	if keyArg == nil {
		return nil, fmt.Errorf("decrypt() requires a key argument")
	}

	// Try to decode as base64 first (standard format)
	keyStr := ""
	if str, ok := keyArg.(string); ok {
		keyStr = str
	} else {
		keyStr = fmt.Sprintf("%v", keyArg)
	}

	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("decrypt() key must be base64-encoded: %v", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("decrypt() key must decode to 32 bytes for AES-256 (got %d)", len(keyBytes))
	}

	// Create AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("decrypt() failed to create cipher: %v", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt() failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("decrypt() ciphertext too short (expected at least %d bytes, got %d)", nonceSize, len(encryptedData))
	}

	// Extract nonce and ciphertext
	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt() failed to decrypt: authentication tag mismatch or corrupted data")
	}

	// Gob-decode to restore original type and structure
	result, err := gobDecode(plaintext)
	if err != nil {
		return nil, fmt.Errorf("decrypt() failed to decode data: %v", err)
	}

	return result, nil
}

// gobEncode encodes a value using gob, handling duso types
func gobEncode(val any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	// Convert duso types to gob-compatible types
	gobVal := toGobValue(val)

	// Wrap in EncryptedValue for proper gob encoding
	wrapped := EncryptedValue{Data: gobVal}

	if err := encoder.Encode(wrapped); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// gobDecode decodes gob-encoded data back to duso values
func gobDecode(data []byte) (any, error) {
	var wrapped EncryptedValue
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&wrapped); err != nil {
		return nil, err
	}
	return wrapped.Data, nil
}

// toGobValue converts duso types to gob-compatible types
func toGobValue(val any) any {
	switch v := val.(type) {
	case nil:
		return nil
	case bool:
		return v
	case float64:
		return v
	case string:
		return v
	case []any:
		// Array: convert each element
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = toGobValue(item)
		}
		return result
	case *[]Value:
		// Duso array ref: convert to []any
		result := make([]any, len(*v))
		for i, item := range *v {
			result[i] = toGobValue(ValueToInterface(item))
		}
		return result
	case map[string]any:
		// Object: convert recursively
		result := make(map[string]any)
		for k, item := range v {
			result[k] = toGobValue(item)
		}
		return result
	case *ValueRef:
		// Handle wrapped values (binary, code, error, etc.)
		if v.Val.IsBinary() {
			binVal := v.Val.AsBinary()
			if binVal != nil && binVal.Data != nil {
				return *binVal.Data
			}
		}
		// For other wrapped types, stringify
		return script.ValueToDusoString(v.Val)
	default:
		// Fallback: try to use as-is
		return v
	}
}

// toBytes converts various types to []byte using duso string representation
func toBytes(val any) []byte {
	// Handle binary data
	if v, ok := val.(script.Value); ok && v.IsBinary() {
		binVal := v.AsBinary()
		if binVal != nil && binVal.Data != nil {
			return *binVal.Data
		}
	}
	if v, ok := val.(*script.ValueRef); ok && v.Val.IsBinary() {
		binVal := v.Val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			return *binVal.Data
		}
	}

	// Handle string
	if str, ok := val.(string); ok {
		return []byte(str)
	}

	// Use duso's string representation for all other types
	scriptVal := InterfaceToValue(val)
	dusoStr := script.ValueToDusoString(scriptVal)
	return []byte(dusoStr)
}
