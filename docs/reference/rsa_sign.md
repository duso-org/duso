# rsa_sign()

Sign data with an RSA private key using SHA256-PKCS1v15.

## Signature

```duso
rsa_sign(data, private_key_pem)
```

## Parameters

- `data` (string | binary) - The data to sign
- `private_key_pem` (string) - PEM-encoded RSA private key (PKCS1 or PKCS8 format)

## Returns

Base64-encoded signature string

## Examples

Sign a string with an RSA private key:

```duso
private_key = load("/path/to/private_key.pem")
data = "message to sign"
signature = rsa_sign(data, private_key)
print(signature)  // Base64-encoded signature
```

Sign binary data (e.g., file):

```duso
private_key = load("/path/to/private_key.pem")
file_data = load_binary("document.pdf")
signature = rsa_sign(file_data, private_key)
print("File signed: " + signature)
```

Code signing workflow:

```duso
private_key = load("code_signing_key.pem")
code = load("release-1.0.du")
code_signature = rsa_sign(code, private_key)
save("release-1.0.du.sig", code_signature)
```

## Key Format

Accepts PEM-encoded RSA private keys in these formats:

```
-----BEGIN RSA PRIVATE KEY-----
[base64 encoded PKCS1 key]
-----END RSA PRIVATE KEY-----
```

Or PKCS8 format:

```
-----BEGIN PRIVATE KEY-----
[base64 encoded PKCS8 key]
-----END PRIVATE KEY-----
```

Generate a test key with OpenSSL:

```bash
openssl genrsa -out private_key.pem 2048
```

## Security Notes

- Uses SHA256 with PKCS1v15 padding (RSA-SHA256)
- Requires valid RSA private key in PEM format
- Signature can be verified with `rsa_verify()` using the corresponding public key
- Never share or expose private keys
- Different signatures are generated each time (randomized PKCS1v15 padding)

## Common Use Cases

- **Code signing**: Sign code releases for integrity and authenticity
- **API authentication**: Sign requests for server verification
- **Digital signatures**: Sign documents for non-repudiation
- **JWT tokens**: Create signed tokens with RSA keys
- **Certificate-based auth**: Sign authentication proofs

## Errors

Throws an error if:

- Data is empty
- Private key PEM is invalid or unparseable
- Key is not an RSA key (e.g., EC or DSA)
- Key is malformed or corrupted

## See Also

- [rsa_verify() - Verify RSA signatures](/docs/reference/rsa_verify.md)
- [hash() - Compute cryptographic hashes](/docs/reference/hash.md)
- [encode_base64() - Base64 encoding](/docs/reference/encode_base64.md)
