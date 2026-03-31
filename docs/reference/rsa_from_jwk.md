# rsa_from_jwk()

Convert JWK (JSON Web Key) components to a PEM-encoded RSA public key.

## Signature

```duso
rsa_from_jwk(n, e)
```

## Parameters

- `n` (string) - Base64url-encoded RSA modulus (from JWK)
- `e` (string) - Base64url-encoded RSA public exponent (from JWK, typically "AQAB")

## Returns

PEM-encoded RSA public key string (ready for use with `verify_rsa()`)

## Examples

Convert Apple's JWK to PEM for token verification:

```duso
// Apple's public key components (from https://appleid.apple.com/auth/keys)
n = "vcDUGnc9ITh348cRCn6CENlcFzOm4X_sxDyPumPZrM3YhH_zXfjNhBCQnvTGNFqGzsqok87ufbWSEqYiYQDsh8DMTT_tx5bcuRJI-LmuX3CkLOKq0KXVUzijpj45mTvdGoC_dL2ei_nGs9yz0EJwilNpwPZxkGxNhWi7MWobOd4BjzBIkqDw_HqKZ_486EKHhyV0qgXfwQYgnKT9blBYc6ZNej9MPHyve5lZs084uEiY_UYjV0rlxfZdYa0g3scG7wc2dWMlqZ4QvbPMj0KTzMNtO-9cr3aruTTPQ2qDqFAThZDNrPaScJIXAcgrARvqy1CAMT_8gSYFbb4Ld0tRbQ"
e = "AQAB"

public_key = rsa_from_jwk(n, e)
print(public_key)  // PEM-formatted public key
```

Verify a signed token:

```duso
// Fetch Apple's keys
apple_keys = fetch("https://appleid.apple.com/auth/keys").json()

// Find the key by kid (key ID from token header)
key = nil
for k in apple_keys.keys do
  if k.kid == token_kid then
    key = k
    break
  end
end

// Convert JWK to PEM
public_key = rsa_from_jwk(key.n, key.e)

// Verify the token signature
if verify_rsa(token_payload, token_signature, public_key) then
  print("✓ Token signature verified")
else
  print("✗ Token signature invalid")
end
```

Complete OAuth OIDC token verification flow:

```duso
// 1. Parse JWT (decode header, claims)
parts = split(token, ".")
header = parse_json(decode_base64(parts[0]))
claims = parse_json(decode_base64(parts[1]))
signature = parts[2]

// 2. Fetch provider's public keys
provider_jwks = fetch(jwks_uri).json()

// 3. Find matching key by kid
key = nil
for k in provider_jwks.keys do
  if k.kid == header.kid then
    key = k
    break
  end
end

if not key then
  throw("Key not found: " + header.kid)
end

// 4. Convert JWK to PEM
public_key = rsa_from_jwk(key.n, key.e)

// 5. Verify signature
token_to_verify = parts[0] + "." + parts[1]
if verify_rsa(token_to_verify, signature, public_key) then
  print("✓ Token verified, claims: " + format_json(claims))
else
  throw("Token signature invalid")
end
```

## Common JWK Sources

**Apple Sign In:** `https://appleid.apple.com/auth/keys`
```duso
keys = fetch("https://appleid.apple.com/auth/keys").json()
// Returns: { keys: [ { kty: "RSA", kid: "...", n: "...", e: "AQAB" }, ... ] }
```

**Google OAuth:** `https://www.googleapis.com/oauth2/v3/certs`
```duso
keys = fetch("https://www.googleapis.com/oauth2/v3/certs").json()
// Returns: { keys: [ { kty: "RSA", kid: "...", n: "...", e: "AQAB" }, ... ] }
```

**Auth0:** `https://{domain}/.well-known/jwks.json`
```duso
domain = "your-tenant.auth0.com"
keys = fetch("https://" + domain + "/.well-known/jwks.json").json()
```

## Use Cases

- **OAuth/OIDC verification**: Verify JWT tokens from Apple Sign In, Google OAuth, Auth0, etc.
- **Token signature validation**: Verify RS256-signed tokens without trusting the client
- **Key rotation**: Fetch provider's latest public keys and dynamically convert them
- **Multi-key setup**: Handle providers that rotate keys via `kid` matching

## Key Format

Input: Base64url-encoded JWK components from a JSON Web Key Set

```json
{
  "kty": "RSA",
  "kid": "key-id-123",
  "use": "sig",
  "alg": "RS256",
  "n": "base64url_encoded_modulus",
  "e": "AQAB"
}
```

Output: PEM-encoded RSA public key string

```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
...base64_encoded_key_material...
-----END PUBLIC KEY-----
```

## Security Notes

- JWK `n` and `e` components are **public data** (found on provider endpoints)
- The output PEM is a **public key** (can be freely shared)
- Always verify `alg` field is "RS256" before verification (don't accept other algorithms)
- Match `kid` in JWT header to `kid` in JWKS before converting (prevents key confusion)
- Cache fetched JWKS locally with reasonable TTL (e.g., 24 hours) to avoid repeated network calls
- Validate `iss` (issuer), `aud` (audience), and `exp` (expiration) claims after signature verification

## Errors

Throws an error if:

- `n` is missing or not a valid base64url string
- `e` is missing or not a valid base64url string
- Base64url decoding fails
- Key material is invalid

## See Also

- [verify_rsa() - Verify RSA signatures](/docs/reference/verify_rsa.md)
- [sign_rsa() - Create RSA signatures](/docs/reference/sign_rsa.md)
- [decode_base64() - Decode base64 strings](/docs/reference/decode_base64.md)
- [RFC 7517 - JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
- [RFC 7518 - JSON Web Algorithms (JWA)](https://tools.ietf.org/html/rfc7518)
