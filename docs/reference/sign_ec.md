# sign_ec()

Sign data with an EC (elliptic curve) private key using ES256 (ECDSA with SHA256).

## Signature

```duso
sign_ec(data, private_key_pem)
```

## Parameters

- `data` (string | binary) - The data to sign
- `private_key_pem` (string) - PEM-encoded EC private key (P-256 curve)

## Returns

Base64url-encoded signature string (IEEE P1363 format: raw r||s, 64 bytes for P-256)

## Examples

Sign a string with an EC private key:

```duso
private_key = load("/path/to/ec_private_key.pem")
data = "message to sign"
signature = sign_ec(data, private_key)
print(signature)  // Base64-encoded ES256 signature
```

Sign binary data:

```duso
private_key = load("/path/to/ec_private_key.pem")
file_data = load_binary("document.pdf")
signature = sign_ec(file_data, private_key)
print("File signed: " + signature)
```

## Key Format

Accepts PEM-encoded EC private keys (P-256 curve):

```
-----BEGIN EC PRIVATE KEY-----
[base64 encoded EC private key]
-----END EC PRIVATE KEY-----
```

Or PKCS8 format:

```
-----BEGIN PRIVATE KEY-----
[base64 encoded PKCS8 EC private key]
-----END PRIVATE KEY-----
```

Generate a test key with OpenSSL:

```bash
openssl ecparam -name prime256v1 -genkey -noout -out private_key.pem
```

## Signature Format

The signature is returned as **base64url(r||s)** in IEEE P1363 format:
- `r` and `s` are each 32 bytes (P-256 coordinates)
- Total signature is 64 bytes: 32 bytes r + 32 bytes s
- This is the standard JWT/OIDC format for ES256 signatures
- Compatible with App Store Server API, OIDC providers, and all JWT libraries

## Security Notes

- Uses SHA256 with ECDSA (ES256)
- Requires valid EC private key in PEM format for P-256 curve
- Signature can be verified with `verify_ec()` using the corresponding public key
- Different signatures are generated each time (randomized ECDSA nonce)
- Never share or expose private keys

## Common Use Cases

- **API authentication**: Sign requests with EC keys
- **Digital signatures**: Sign documents for non-repudiation
- **Webhook verification**: Sign outgoing webhooks
- **Certificate-based auth**: Sign authentication proofs
- **Token signing**: Create ES256-signed JWT tokens

## Errors

Throws an error if:

- Data is empty
- Private key PEM is invalid or unparseable
- Key is not an EC key (e.g., RSA or DSA)
- Key is not P-256 curve
- Key is malformed or corrupted

## See Also

- [verify_ec() - Verify EC signatures](/docs/reference/verify_ec.md)
- [ec_from_jwk() - Convert JWK to EC public key](/docs/reference/ec_from_jwk.md)
- [sign_rsa() - RSA signing](/docs/reference/sign_rsa.md)
- [hash() - Compute cryptographic hashes](/docs/reference/hash.md)
