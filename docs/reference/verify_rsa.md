# verify_rsa()

Verify an RSA signature using SHA256-PKCS1v15.

## Signature

```duso
verify_rsa(data, signature, public_key_pem)
```

## Parameters

- `data` (string | binary) - The data that was signed
- `signature` (string) - Base64-encoded signature (from `rsa_sign()`)
- `public_key_pem` (string) - PEM-encoded RSA public key or X.509 certificate

## Returns

Boolean: `true` if signature is valid, `false` if invalid (never throws on verification failure)

## Examples

Verify a signature with a public key:

```duso
public_key = load("/path/to/public_key.pem")
data = "message to verify"
signature = load("message.sig")
is_valid = verify_rsa(data, signature, public_key)
print("Signature valid: " + tostring(is_valid))
```

Verify signed binary data (e.g., file):

```duso
public_key = load("/path/to/public_key.pem")
file_data = load_binary("document.pdf")
file_sig = load("document.pdf.sig")
if verify_rsa(file_data, file_sig, public_key) then
  print("File signature verified")
else
  print("File signature invalid - data may be tampered")
end
```

Code verification workflow:

```duso
public_key = load("code_signing_key.pub")
code = load("release-1.0.du")
code_sig = load("release-1.0.du.sig")

if verify_rsa(code, code_sig, public_key) then
  print("✓ Code signature valid, safe to run")
  eval(parse(code))
else
  print("✗ Code signature invalid - REJECTED")
end
```

Verify with X.509 certificate:

```duso
// Can use certificate directly instead of extracting public key
cert_pem = load("server.crt")
data = "authenticated request"
signature = load("request.sig")

if verify_rsa(data, signature, cert_pem) then
  print("Certificate-based signature verified")
end
```

## Key Format

Accepts PEM-encoded RSA public keys:

```
-----BEGIN PUBLIC KEY-----
[base64 encoded PKIX key]
-----END PUBLIC KEY-----
```

Or X.509 certificates:

```
-----BEGIN CERTIFICATE-----
[base64 encoded certificate]
-----END CERTIFICATE-----
```

Extract public key from private key with OpenSSL:

```bash
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

## Security Notes

- Uses SHA256 with PKCS1v15 padding (RSA-SHA256), matching `rsa_sign()`
- Returns `false` instead of throwing on verification failure
- Safe to use in conditionals without try/catch for invalid signatures
- Only throws on PEM parsing errors or missing/invalid parameters
- Public keys can be safely shared - only private keys need protection

## Common Use Cases

- **Code signing verification**: Verify released code integrity
- **API authentication**: Verify signed requests
- **Digital signatures**: Verify signed documents
- **JWT verification**: Verify RSA-signed tokens
- **Certificate-based auth**: Verify identity based on certs
- **Supply chain security**: Verify integrity of dependencies

## Signature Failure vs Errors

**Returns false** (signature doesn't match):
- Data was modified
- Wrong signature provided
- Wrong public key used
- Original signature was tampered with

**Throws error** (validation error):
- Public key PEM is invalid or unparseable
- Key is not an RSA key
- Signature is not valid base64
- Parameters missing

## See Also

- [rsa_sign() - Create RSA signatures](/docs/reference/rsa_sign.md)
- [hash() - Compute cryptographic hashes](/docs/reference/hash.md)
- [decode_base64() - Base64 decoding](/docs/reference/decode_base64.md)
