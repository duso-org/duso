# decrypt()

Decrypt data that was encrypted with `encrypt()` using AES-256-GCM. Automatically restores the original type (string, number, array, object, etc.).

`decrypt(encrypted_data, key)`

## Parameters

- `encrypted_data` (string) - Base64-encoded encrypted data (output from `encrypt()`)
- `key` (string) - Base64-encoded 32-byte encryption key. Must be the same key used to encrypt the data

## Returns

The original decrypted value with original type preserved (string, number, array, object, boolean, etc.)

## Errors

Throws an error if:
- `key` is not exactly 32 bytes
- `encrypted_data` is not valid base64
- `encrypted_data` is corrupted or tampered with (authentication tag verification fails)
- `encrypted_data` was encrypted with a different key

## Examples

Decrypt a string:

```duso
key_b64 = encode_base64("01234567890123456789012345678901")
encrypted = "L0F5K+L2D..."  // Base64-encoded ciphertext from encrypt()

decrypted = decrypt(encrypted, key_b64)
print(decrypted)  // Original plaintext string
```

Decrypt an object with type preservation:

```duso
key_b64 = encode_base64("0123456789abcdef0123456789abcdef")
encrypted = encrypt({name = "alice", age = 30}, key_b64)

decrypted = decrypt(encrypted, key_b64)
print(type(decrypted))          // "object"
print(decrypted.name)           // "alice"
print(decrypted.age)            // 30
```

Retrieve and decrypt from datastore:

```duso
key = env("CRYPTO_KEY")  // Base64-encoded
ds = datastore("users")

// Store encrypted sensitive data
private_data = {email = "alice@example.com", ssn = "123-45-6789"}
ds.set("user:123", {name = "alice", _private = encrypt(private_data, key)})

// Later, retrieve and decrypt
user = ds.get("user:123")
private = decrypt(user._private, key)
print("Email: " + private.email)
```

Decrypt an array:

```duso
key = "01234567890123456789012345678901"
original_array = [1, 2, 3, 4, 5]
encrypted = encrypt(original_array, key)

decrypted = decrypt(encrypted, key)
print(type(decrypted))       // "array"
print(len(decrypted))        // 5
print(decrypted[0])          // 1
```

Handle decryption errors:

```duso
key = "01234567890123456789012345678901"
wrong_key = "wrongkey1234567890123456789012"

try
  encrypted = encrypt("secret", key)
  decrypted = decrypt(encrypted, wrong_key)  // Wrong key!
catch (err)
  print("Decryption failed: " + err)  // "authentication tag mismatch or corrupted data"
end
```

## Type Restoration

`decrypt()` automatically restores the original Duso type:

```duso
key_b64 = encode_base64("01234567890123456789012345678901")

// String
enc_str = encrypt("text", key_b64)
dec_str = decrypt(enc_str, key_b64)
// type(dec_str) == "string" ✓

// Number
enc_num = encrypt(42.5, key_b64)
dec_num = decrypt(enc_num, key_b64)
// type(dec_num) == "number" ✓

// Array
enc_arr = encrypt([1, 2, 3], key_b64)
dec_arr = decrypt(enc_arr, key_b64)
// type(dec_arr) == "array" ✓

// Object
enc_obj = encrypt({a = 1, b = 2}, key_b64)
dec_obj = decrypt(enc_obj, key_b64)
// type(dec_obj) == "object" ✓

// Boolean
enc_bool = encrypt(true, key_b64)
dec_bool = decrypt(enc_bool, key_b64)
// type(dec_bool) == "boolean" ✓
```

## Performance

- Decryption is CPU-bound (< 1 microsecond for small values on modern CPUs)
- Hardware-accelerated on CPUs with AES-NI support
- Authentication tag verification is constant-time (resistant to timing attacks)
- Type restoration via gob decoding adds small overhead

## Key Management

The decryption key must:
- Be **base64-encoded 32-byte key** (same format as used by `encrypt()` and datastore)
- Be the **same key used for encryption** (exact match)
- Be **kept confidential** (stored in environment variables, not source code)

```duso
// ✓ Good: Load base64-encoded key from environment
key_b64 = env("DATA_ENCRYPTION_KEY")
decrypted = decrypt(encrypted, key_b64)

// ✗ Bad: Hardcoded key or raw bytes
key_b64 = "01234567890123456789012345678901"  // Will fail - must be base64
```

## Security Notes

- **Authenticity**: Decryption fails if ciphertext is corrupted or tampered with (GCM detects this)
- **Confidentiality**: Ensures original data remains secret (anyone with the key can decrypt)
- **Nonce**: Each encryption uses a unique random nonce; safe to encrypt multiple values with same key
- **Key leakage**: If the key is compromised, all encrypted data is compromised
- **Constant-time**: Authentication tag verification runs in constant time (no timing side-channel)

## Errors and Troubleshooting

**"key must be base64-encoded"**
- Key is not valid base64
- Key should be passed as `env("DATA_ENCRYPTION_KEY")` or `encode_base64("raw-32-byte-key")`

**"key must decode to 32 bytes"**
- Key decodes to wrong length
- Generate a new key: `openssl rand 32 | base64`
- Verify: `len(decode_base64(key_b64))` should equal 32

**"authentication tag mismatch or corrupted data"**
- Wrong key was used
- Ciphertext was modified or corrupted
- Ciphertext was not produced by `encrypt()`

**"ciphertext too short"**
- Invalid base64 or truncated encrypted data
- Ciphertext was not produced by `encrypt()`

## See Also

- [encrypt() - Encrypt with AES-256-GCM](/docs/reference/encrypt.md)
- [datastore() encryption at rest - Transparent disk encryption](/docs/reference/datastore.md#encryption-at-rest)
- [encode_base64() - Base64 encoding](/docs/reference/encode_base64.md)
- [hash() - Cryptographic hashing](/docs/reference/hash.md)
