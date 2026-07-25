# encrypt()

Encrypt data using AES-256-GCM symmetric encryption. Encrypts any Duso value (string, number, array, object, etc.) with automatic type preservation for later decryption.

`encrypt(data, key)`

## Parameters

- `data` - Any Duso value to encrypt (string, number, array, object, boolean, etc.). Values are gob-encoded before encryption to preserve type information
- `key` (string) - Base64-encoded 32-byte encryption key (for AES-256). Use `encode_base64()` to encode a raw 32-byte key, or load from environment via `env()`

## Returns

Base64-encoded encrypted data (string). Contains a random nonce and authentication tag. Same data encrypted twice produces different ciphertexts due to random nonce.

## Examples

Encrypt and decrypt a string:

```duso
key_b64 = encode_base64("01234567890123456789012345678901")  // 32 bytes encoded as base64
plaintext = "secret message"
encrypted = encrypt(plaintext, key_b64)
print(encrypted)  // Base64-encoded ciphertext

decrypted = decrypt(encrypted, key_b64)
print(decrypted)  // "secret message"
```

Encrypt an object with sensitive fields:

```duso
key_b64 = encode_base64("0123456789abcdef0123456789abcdef")
user = {
  id = 123,
  name = "alice",
  email = "alice@example.com",
  phone = "555-1234"
}

encrypted = encrypt(user, key_b64)
// Later, decrypt to retrieve full object
user_decrypted = decrypt(encrypted, key_b64)
print(user_decrypted.email)  // "alice@example.com"
```

Store encrypted values in datastore:

```duso
key = env("CRYPTO_KEY")  // Base64-encoded from environment
ds = datastore("users")

// Group sensitive fields and encrypt once for efficiency
private_data = {
  email = "alice@example.com",
  phone = "555-1234",
  ssn = "123-45-6789"
}

user_record = {
  id = 123,
  name = "alice",
  role = "admin",
  _encrypted = encrypt(private_data, key)
}

ds.set("user:123", user_record)

// Later, retrieve and decrypt
user = ds.get("user:123")
private = decrypt(user._encrypted, key)
print("Email: " + private.email)
```

Encrypt an array of values:

```duso
key = "01234567890123456789012345678901"
items = ["item1", "item2", "item3"]
encrypted_array = encrypt(items, key)

decrypted_array = decrypt(encrypted_array, key)
print(len(decrypted_array))  // 3
print(decrypted_array[0])    // "item1"
```

## Algorithm Notes

- **Algorithm**: AES-256-GCM (Advanced Encryption Standard with Galois/Counter Mode)
- **Key size**: Exactly 32 bytes (256 bits)
- **Nonce**: 12 bytes, randomly generated per encryption
- **Authentication tag**: Included in ciphertext to prevent tampering
- **Output format**: Base64-encoded string containing nonce + ciphertext + tag

## Type Preservation

Unlike string-based encoding, `encrypt()` preserves Duso types through gob serialization:

```duso
key_b64 = encode_base64("01234567890123456789012345678901")

// Array
arr = [1, 2, 3]
encrypted_arr = encrypt(arr, key_b64)
decrypted_arr = decrypt(encrypted_arr, key_b64)
// type(decrypted_arr) == "array" ✓

// Object
obj = {name = "alice", age = 30}
encrypted_obj = encrypt(obj, key_b64)
decrypted_obj = decrypt(encrypted_obj, key_b64)
// type(decrypted_obj) == "object" ✓
// decrypted_obj.name == "alice" ✓

// Number
num = 42
encrypted_num = encrypt(num, key_b64)
decrypted_num = decrypt(encrypted_num, key_b64)
// type(decrypted_num) == "number" ✓
// decrypted_num == 42 ✓
```

## Key Management

Best practices for encryption keys:

- **Generate**: Use a cryptographically secure random 32-byte key, then base64-encode it (e.g., `openssl rand 32 | base64`)
- **Storage**: Store keys in environment variables, not in code or config files
- **Format**: Keys must be base64-encoded 32-byte strings (same standard as datastore)
- **Rotation**: Plan for key rotation; old data encrypted with old keys remains encrypted
- **Access**: Limit key access to authorized services and admins only

```duso
// ✓ Good: Load base64-encoded key from environment
key_b64 = env("DATA_ENCRYPTION_KEY")
encrypted = encrypt(data, key_b64)

// ✗ Bad: Key in source code
key_b64 = encode_base64("01234567890123456789012345678901")

// ✗ Bad: Key not base64-encoded
key = "01234567890123456789012345678901"  // Will fail - must be base64
```

## Performance

- Encryption/decryption is CPU-bound and fast (typically < 1 microsecond for small values)
- AES-256 is hardware-accelerated on modern CPUs (AES-NI)
- Type preservation via gob adds small overhead (serialization cost)
- For best efficiency: encrypt related fields together rather than individually

## Common Use Cases

- **Sensitive datastore fields**: Encrypt email, phone, SSN before storing
- **API request/response encryption**: Encrypt payloads in untrusted networks
- **Secure configuration**: Encrypt secrets stored in files or databases
- **End-to-end encryption**: Share encrypted data with other services/users

## Security Notes

- **Confidentiality**: AES-256-GCM provides strong encryption (not breakable by brute force)
- **Integrity**: GCM mode provides authentication; detects tampering automatically
- **Nonce**: Randomly generated per encryption; never reused with the same key
- **Key reuse**: Same key can encrypt multiple values (nonce prevents attacks)
- **Timing attacks**: GCM authentication is constant-time (resistant to timing attacks)

## Differences from hash()

- `hash()` is one-way; cannot recover original data
- `encrypt()` is reversible (with correct key)
- `hash()` for data integrity verification
- `encrypt()` for data confidentiality (privacy)

## See Also

- [decrypt() - Decrypt AES-256-GCM encrypted data](/docs/reference/decrypt.md)
- [datastore() encryption at rest - Transparent disk encryption](/docs/reference/datastore.md#encryption-at-rest)
- [hash() - Cryptographic hashing](/docs/reference/hash.md)
- [hmac() - Message authentication](/docs/reference/hmac.md)
- [encode_base64() - Base64 encoding](/docs/reference/encode_base64.md)
