package runtime

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/duso-org/duso/pkg/script"
)

// builtinSignRSA signs data with an RSA private key
// Usage: sign_rsa(data, private_key_pem)
// Returns: base64-encoded signature
func builtinSignRSA(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get data - support both positional (0) and named (data)
	var dataBytes []byte
	var dataArg any

	if d, ok := args["data"]; ok {
		dataArg = d
	} else if d, ok := args["0"]; ok {
		dataArg = d
	}

	if dataArg == nil {
		return nil, fmt.Errorf("sign_rsa() requires a data argument")
	}

	// Handle binary data
	if val, ok := dataArg.(script.Value); ok && val.IsBinary() {
		binVal := val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			dataBytes = *binVal.Data
		}
	} else if val, ok := dataArg.(*script.ValueRef); ok && val.Val.IsBinary() {
		binVal := val.Val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			dataBytes = *binVal.Data
		}
	} else if str, ok := dataArg.(string); ok {
		dataBytes = []byte(str)
	} else {
		dataBytes = []byte(fmt.Sprintf("%v", dataArg))
	}

	if len(dataBytes) == 0 {
		return nil, fmt.Errorf("sign_rsa() requires non-empty data")
	}

	// Get private key PEM - support both positional (1) and named (private_key_pem)
	var keyPEM string
	if key, ok := args["private_key_pem"]; ok {
		if keyStr, ok := key.(string); ok {
			keyPEM = keyStr
		}
	} else if key, ok := args["1"]; ok {
		if keyStr, ok := key.(string); ok {
			keyPEM = keyStr
		}
	}

	if keyPEM == "" {
		return nil, fmt.Errorf("sign_rsa() requires a private_key_pem string argument")
	}

	// Parse the PEM-encoded private key
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, fmt.Errorf("sign_rsa() failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		parsedKey, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("sign_rsa() failed to parse private key: %v", err)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("sign_rsa() key is not an RSA private key")
		}
	}

	// Sign the data using SHA256
	hash := sha256.Sum256(dataBytes)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign_rsa() failed to sign data: %v", err)
	}

	// Return base64-encoded signature
	return base64.StdEncoding.EncodeToString(signature), nil
}

// builtinVerifyRSA verifies an RSA signature
// Usage: verify_rsa(data, signature, public_key_pem)
// Returns: true if signature is valid, false otherwise
func builtinVerifyRSA(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get data - support both positional (0) and named (data)
	var dataBytes []byte
	var dataArg any

	if d, ok := args["data"]; ok {
		dataArg = d
	} else if d, ok := args["0"]; ok {
		dataArg = d
	}

	if dataArg == nil {
		return nil, fmt.Errorf("verify_rsa() requires a data argument")
	}

	// Handle binary data
	if val, ok := dataArg.(script.Value); ok && val.IsBinary() {
		binVal := val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			dataBytes = *binVal.Data
		}
	} else if val, ok := dataArg.(*script.ValueRef); ok && val.Val.IsBinary() {
		binVal := val.Val.AsBinary()
		if binVal != nil && binVal.Data != nil {
			dataBytes = *binVal.Data
		}
	} else if str, ok := dataArg.(string); ok {
		dataBytes = []byte(str)
	} else {
		dataBytes = []byte(fmt.Sprintf("%v", dataArg))
	}

	if len(dataBytes) == 0 {
		return nil, fmt.Errorf("verify_rsa() requires non-empty data")
	}

	// Get signature (base64-encoded) - support both positional (1) and named (signature)
	var signatureB64 string
	if sig, ok := args["signature"]; ok {
		if sigStr, ok := sig.(string); ok {
			signatureB64 = sigStr
		}
	} else if sig, ok := args["1"]; ok {
		if sigStr, ok := sig.(string); ok {
			signatureB64 = sigStr
		}
	}

	if signatureB64 == "" {
		return nil, fmt.Errorf("verify_rsa() requires a signature string argument")
	}

	// Decode base64 signature
	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("verify_rsa() failed to decode signature: %v", err)
	}

	// Get public key PEM - support both positional (2) and named (public_key_pem)
	var keyPEM string
	if key, ok := args["public_key_pem"]; ok {
		if keyStr, ok := key.(string); ok {
			keyPEM = keyStr
		}
	} else if key, ok := args["2"]; ok {
		if keyStr, ok := key.(string); ok {
			keyPEM = keyStr
		}
	}

	if keyPEM == "" {
		return nil, fmt.Errorf("verify_rsa() requires a public_key_pem string argument")
	}

	// Parse the PEM-encoded public key
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, fmt.Errorf("verify_rsa() failed to parse PEM block")
	}

	var publicKey *rsa.PublicKey

	// Try to parse as a public key
	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		// Try to parse as a certificate
		cert, err2 := x509.ParseCertificate(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("verify_rsa() failed to parse public key: %v", err)
		}
		var ok bool
		publicKey, ok = cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("verify_rsa() certificate public key is not an RSA key")
		}
	} else {
		var ok bool
		publicKey, ok = pubKeyInterface.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("verify_rsa() public key is not an RSA key")
		}
	}

	// Verify the signature
	hash := sha256.Sum256(dataBytes)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		// Signature is invalid - return false, don't throw
		return false, nil
	}

	return true, nil
}
