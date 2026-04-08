package runtime

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"

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

	// Try to parse as a certificate first (some providers send certificates instead of bare keys)
	cert, certErr := x509.ParseCertificate(block.Bytes)
	if certErr == nil {
		// Successfully parsed as certificate
		var ok bool
		publicKey, ok = cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("verify_rsa() certificate public key is not an RSA key")
		}
	} else {
		// Try to parse as a bare public key
		pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("verify_rsa() failed to parse public key: %v", err)
		}
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

// builtinRSAFromJWK converts JWK (modulus and exponent) to PEM-encoded RSA public key
// Usage: rsa_from_jwk(n, e)
// where n and e are base64url-encoded JWK components
// Returns: PEM-encoded RSA public key string
func builtinRSAFromJWK(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get n (modulus) - support both positional (0) and named (n)
	var nStr string
	if n, ok := args["n"]; ok {
		if nVal, ok := n.(string); ok {
			nStr = nVal
		}
	} else if n, ok := args["0"]; ok {
		if nVal, ok := n.(string); ok {
			nStr = nVal
		}
	}

	if nStr == "" {
		return nil, fmt.Errorf("rsa_from_jwk() requires an 'n' (modulus) argument")
	}

	// Get e (exponent) - support both positional (1) and named (e)
	var eStr string
	if e, ok := args["e"]; ok {
		if eVal, ok := e.(string); ok {
			eStr = eVal
		}
	} else if e, ok := args["1"]; ok {
		if eVal, ok := e.(string); ok {
			eStr = eVal
		}
	}

	if eStr == "" {
		return nil, fmt.Errorf("rsa_from_jwk() requires an 'e' (exponent) argument")
	}

	// Decode base64url modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_from_jwk() failed to decode modulus: %v", err)
	}

	// Decode base64url exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_from_jwk() failed to decode exponent: %v", err)
	}

	// Convert to big.Int
	modulus := new(big.Int).SetBytes(nBytes)
	exponentBytes := big.NewInt(0).SetBytes(eBytes).Uint64()

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponentBytes),
	}

	// Encode to PEM format
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("rsa_from_jwk() failed to marshal public key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	return string(pemBytes), nil
}

// builtinSignEC signs data with an EC private key using ES256 (ECDSA-SHA256)
// Usage: sign_ec(data, private_key_pem)
// Returns: base64-encoded signature (r,s DER-encoded)
func builtinSignEC(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get data - support both positional (0) and named (data)
	var dataBytes []byte
	var dataArg any

	if d, ok := args["data"]; ok {
		dataArg = d
	} else if d, ok := args["0"]; ok {
		dataArg = d
	}

	if dataArg == nil {
		return nil, fmt.Errorf("sign_ec() requires a data argument")
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
		return nil, fmt.Errorf("sign_ec() requires non-empty data")
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
		return nil, fmt.Errorf("sign_ec() requires a private_key_pem string argument")
	}

	// Parse the PEM-encoded private key
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return nil, fmt.Errorf("sign_ec() failed to parse PEM block")
	}

	// Parse the EC private key
	parsedKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("sign_ec() failed to parse EC private key: %v", err)
	}

	// Sign the data using SHA256 (ES256)
	hash := sha256.Sum256(dataBytes)
	r, s, err := ecdsa.Sign(rand.Reader, parsedKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("sign_ec() failed to sign data: %v", err)
	}

	// Encode r and s as raw bytes (IEEE P1363 format: r||s, 32 bytes each for P-256)
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Pad to 32 bytes if needed (P-256 coordinates are 32 bytes)
	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)

	signature := append(rPadded, sPadded...)

	// Return base64url-encoded signature (JWT standard)
	return base64.RawURLEncoding.EncodeToString(signature), nil
}

// builtinVerifyEC verifies an EC signature using ES256 (ECDSA-SHA256)
// Usage: verify_ec(data, signature, public_key_pem)
// Returns: true if signature is valid, false otherwise
func builtinVerifyEC(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get data - support both positional (0) and named (data)
	var dataBytes []byte
	var dataArg any

	if d, ok := args["data"]; ok {
		dataArg = d
	} else if d, ok := args["0"]; ok {
		dataArg = d
	}

	if dataArg == nil {
		return nil, fmt.Errorf("verify_ec() requires a data argument")
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
		return nil, fmt.Errorf("verify_ec() requires non-empty data")
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
		return nil, fmt.Errorf("verify_ec() requires a signature string argument")
	}

	// Decode base64url signature (IEEE P1363 format: r||s, 64 bytes total for P-256)
	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("verify_ec() failed to decode signature: %v", err)
	}

	// P-256 signatures should be exactly 64 bytes (32 bytes r + 32 bytes s)
	if len(signature) != 64 {
		return false, nil // Invalid signature length
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
		return nil, fmt.Errorf("verify_ec() requires a public_key_pem string argument")
	}

	var derBytes []byte
	var publicKey *ecdsa.PublicKey

	// Check if input is PEM-encoded or raw base64-encoded DER
	if strings.Contains(keyPEM, "-----BEGIN") {
		// Parse PEM format
		block, _ := pem.Decode([]byte(keyPEM))
		if block == nil {
			return nil, fmt.Errorf("verify_ec() failed to parse PEM block")
		}
		derBytes = block.Bytes
	} else {
		// Assume raw base64-encoded DER (Apple x5c certificates are standard base64, not base64url)
		decoded, err := base64.StdEncoding.DecodeString(keyPEM)
		if err != nil {
			// Try base64url as fallback
			decoded, err = base64.RawURLEncoding.DecodeString(keyPEM)
			if err != nil {
				return nil, fmt.Errorf("verify_ec() failed to decode certificate: %v", err)
			}
		}
		derBytes = decoded
	}

	// Try to parse as a certificate first (Apple sends certificates, not bare public keys)
	cert, certErr := x509.ParseCertificate(derBytes)
	if certErr == nil {
		// Successfully parsed as certificate
		var ok bool
		publicKey, ok = cert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("verify_ec() certificate public key is not an EC key")
		}
	} else {
		// Try to parse as a bare public key
		pubKeyInterface, err := x509.ParsePKIXPublicKey(derBytes)
		if err != nil {
			return nil, fmt.Errorf("verify_ec() failed to parse public key: %v", err)
		}
		var ok bool
		publicKey, ok = pubKeyInterface.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("verify_ec() public key is not an EC key")
		}
	}

	// Parse raw r||s format (IEEE P1363: 32 bytes r + 32 bytes s for P-256)
	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])

	// Verify the signature
	hash := sha256.Sum256(dataBytes)
	if ecdsa.Verify(publicKey, hash[:], r, s) {
		return true, nil
	}

	return false, nil
}

// builtinECFromJWK converts JWK (x, y coordinates) to PEM-encoded EC public key (P-256)
// Usage: ec_from_jwk(x, y)
// where x and y are base64url-encoded JWK components
// Returns: PEM-encoded EC public key string
func builtinECFromJWK(evaluator *Evaluator, args map[string]any) (any, error) {
	// Get x coordinate - support both positional (0) and named (x)
	var xStr string
	if x, ok := args["x"]; ok {
		if xVal, ok := x.(string); ok {
			xStr = xVal
		}
	} else if x, ok := args["0"]; ok {
		if xVal, ok := x.(string); ok {
			xStr = xVal
		}
	}

	if xStr == "" {
		return nil, fmt.Errorf("ec_from_jwk() requires an 'x' coordinate argument")
	}

	// Get y coordinate - support both positional (1) and named (y)
	var yStr string
	if y, ok := args["y"]; ok {
		if yVal, ok := y.(string); ok {
			yStr = yVal
		}
	} else if y, ok := args["1"]; ok {
		if yVal, ok := y.(string); ok {
			yStr = yVal
		}
	}

	if yStr == "" {
		return nil, fmt.Errorf("ec_from_jwk() requires a 'y' coordinate argument")
	}

	// Decode base64url x coordinate
	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, fmt.Errorf("ec_from_jwk() failed to decode x coordinate: %v", err)
	}

	// Decode base64url y coordinate
	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, fmt.Errorf("ec_from_jwk() failed to decode y coordinate: %v", err)
	}

	// Convert to big.Int
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	// Create EC public key (P-256 curve)
	curve := elliptic.P256()
	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// Verify the point is on the curve
	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("ec_from_jwk() coordinates are not on P-256 curve")
	}

	// Encode to PEM format
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("ec_from_jwk() failed to marshal public key: %v", err)
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	pemBytes := pem.EncodeToMemory(pemBlock)
	return string(pemBytes), nil
}
