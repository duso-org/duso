# ec_from_jwk()

Convert JWK (JSON Web Key) coordinates (x, y) to a PEM-encoded EC public key (P-256 curve).

## Signature

```duso
ec_from_jwk(x, y)
```

## Parameters

- `x` (string) - Base64url-encoded X coordinate (from JWK, P-256 point)
- `y` (string) - Base64url-encoded Y coordinate (from JWK, P-256 point)

## Returns

PEM-encoded EC public key string (ready for use with `verify_ec()`)

## Examples

Convert Apple's EC JWK to PEM for webhook verification:

```duso
// Apple's webhook public key components (from Apple's webhook verification)
x = "YYG...base64url..."
y = "ZZZ...base64url..."

public_key = ec_from_jwk(x, y)
print(public_key)  // PEM-formatted EC public key
```

Verify an Apple webhook signature:

```duso
// 1. Fetch Apple's EC public keys (JWKS endpoint for webhooks)
// Or receive x,y from webhook metadata

x = "YYG...base64url..."
y = "ZZZ...base64url..."

// 2. Convert JWK to PEM
public_key = ec_from_jwk(x, y)

// 3. Verify the webhook signature
signature = load_base64("webhook_signature")
body = load("webhook_body")

if verify_ec(body, signature, public_key) then
  print("✓ Webhook signature verified")
  // Process webhook payload
else
  print("✗ Webhook signature invalid - REJECT")
end
```

Complete webhook verification flow:

```duso
// 1. Parse webhook metadata to extract x, y coordinates
// (Format depends on webhook provider - Apple, Stripe, etc.)

x = webhook_metadata.x
y = webhook_metadata.y

// 2. Convert JWK to PEM
public_key = ec_from_jwk(x, y)

// 3. Reconstruct signature and body from webhook
signature = webhook_headers["signature"]
body = request.body

// 4. Verify ES256 signature
if verify_ec(body, signature, public_key) then
  // Process authenticated webhook
  process_webhook(body)
else
  // Reject unsigned/tampered webhook
  throw("Webhook signature verification failed")
end
```

## Key Format

Input: Base64url-encoded JWK coordinates for P-256 (NIST prime256v1) curve

```json
{
  "kty": "EC",
  "crv": "P-256",
  "x": "base64url_encoded_x_coordinate",
  "y": "base64url_encoded_y_coordinate",
  "use": "sig"
}
```

Output: PEM-encoded EC public key string

```
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
...base64_encoded_key_material...
-----END PUBLIC KEY-----
```

## Use Cases

- **Apple webhook verification**: Verify signatures on Apple's webhook notifications (In-App Purchase events, etc.)
- **OIDC/OAuth**: Verify tokens from EC-based OIDC providers
- **Webhook authentication**: Verify EC-signed webhooks from any provider
- **Key rotation**: Fetch provider's latest EC public keys and dynamically convert them
- **Multi-key setup**: Handle providers that rotate keys via coordinate matching

## Security Notes

- JWK `x` and `y` coordinates are **public data** (found on provider endpoints)
- The output PEM is a **public key** (can be freely shared)
- Always verify `crv` field is "P-256" before verification (don't accept other curves)
- Coordinates must represent a valid point on the P-256 elliptic curve
- Cache fetched JWKS locally with reasonable TTL (e.g., 24 hours) to avoid repeated network calls
- Validate additional claims (timestamps, signatures, etc.) after coordinate verification

## Errors

Throws an error if:

- `x` is missing or not a valid base64url string
- `y` is missing or not a valid base64url string
- Base64url decoding fails
- Coordinates don't represent a valid point on P-256 curve
- Key material is invalid

## Curve Support

Currently supports **P-256** (also known as prime256v1 or secp256r1) elliptic curve only. This is the standard curve for ES256 (ECDSA with SHA256).

## See Also

- [verify_ec() - Verify EC signatures](/docs/reference/verify_ec.md)
- [sign_ec() - Create EC signatures](/docs/reference/sign_ec.md)
- [rsa_from_jwk() - Convert RSA JWK](/docs/reference/rsa_from_jwk.md)
- [decode_base64() - Decode base64 strings](/docs/reference/decode_base64.md)
- [RFC 7517 - JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
- [RFC 8812 - CBOR Object Signing and Encryption (COSE) and JSON Object Signing and Encryption (JOSE) Registrations for Web Authentication (WebAuthn) Algorithms](https://tools.ietf.org/html/rfc8812)
