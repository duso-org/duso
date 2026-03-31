# verify_ec()

Verify an EC (elliptic curve) signature using ES256 (ECDSA with SHA256).

## Signature

```duso
verify_ec(data, signature, public_key_pem)
```

## Parameters

- `data` (string | binary) - The data that was signed
- `signature` (string) - Base64url-encoded signature in IEEE P1363 format (from `sign_ec()`, or raw r||s format)
- `public_key_pem` (string) - PEM-encoded EC public key or X.509 certificate (P-256 curve)

## Returns

Boolean: `true` if signature is valid, `false` if invalid (never throws on verification failure)

## Examples

Verify a signature with a public key:

```duso
public_key = load("/path/to/public_key.pem")
data = "message to verify"
signature = load("message.sig")
is_valid = verify_ec(data, signature, public_key)
print("Signature valid: " + tostring(is_valid))
```

Verify signed binary data:

```duso
public_key = load("/path/to/public_key.pem")
file_data = load_binary("document.pdf")
file_sig = load("document.pdf.sig")
if verify_ec(file_data, file_sig, public_key) then
  print("File signature verified")
else
  print("File signature invalid - data may be tampered")
end
```

Verify with X.509 certificate:

```duso
// Can use certificate directly instead of extracting public key
cert_pem = load("server.crt")
data = "authenticated request"
signature = load("request.sig")

if verify_ec(data, signature, cert_pem) then
  print("Certificate-based signature verified")
end
```

## Key Format

Accepts PEM-encoded EC public keys (P-256 curve):

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
openssl ec -in private_key.pem -pubout -out public_key.pem
```

## Security Notes

- Uses SHA256 with ECDSA (ES256), matching `sign_ec()`
- Returns `false` instead of throwing on verification failure
- Safe to use in conditionals without try/catch for invalid signatures
- Only throws on PEM parsing errors or missing/invalid parameters
- Public keys can be safely shared - only private keys need protection

## Common Use Cases

- **Webhook verification**: Verify incoming webhook signatures
- **JWT verification**: Verify ES256-signed tokens
- **Request authentication**: Verify EC-signed API requests
- **Digital signatures**: Verify signed documents
- **Certificate-based auth**: Verify identity based on certs
- **Apple webhook notifications**: Verify Apple's EC-signed webhooks

## Signature Failure vs Errors

**Returns false** (signature doesn't match):
- Data was modified
- Wrong signature provided
- Wrong public key used
- Original signature was tampered with

**Throws error** (validation error):
- Public key PEM is invalid or unparseable
- Key is not an EC key
- Signature is not valid base64
- Parameters missing

## See Also

- [sign_ec() - Create EC signatures](/docs/reference/sign_ec.md)
- [ec_from_jwk() - Convert JWK to EC public key](/docs/reference/ec_from_jwk.md)
- [verify_rsa() - RSA verification](/docs/reference/verify_rsa.md)
- [hash() - Compute cryptographic hashes](/docs/reference/hash.md)
