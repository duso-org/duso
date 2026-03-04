# hash()

Compute a cryptographic hash of a string using the specified algorithm.

## Signature

```duso
hash(algo, data)
```

## Parameters

- `algo` (string) - The hash algorithm to use: `"sha256"`, `"sha512"`, `"sha1"`, or `"md5"`
- `data` (string) - The string data to hash

## Returns

Hex-encoded hash string

## Examples

Hash with SHA256 (default recommended):

```duso
hash_value = hash("sha256", "hello world")
print(hash_value)  // "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

Hash with SHA512:

```duso
long_hash = hash("sha512", "password123")
print(long_hash)  // Full 128-character hex string
```

Using named arguments:

```duso
result = hash(algo = "sha256", data = "some data")
print(result)
```

Generate file checksums:

```duso
file_content = load("document.txt")
checksum = hash("sha256", file_content)
save("checksums.txt", "document.txt: " + checksum)
```

Verify data integrity:

```duso
original_data = "important information"
original_hash = hash("sha256", original_data)

// Later, verify the data hasn't changed
current_hash = hash("sha256", original_data)
if original_hash == current_hash {
  print("Data integrity verified")
}
```

## Algorithm Notes

- **sha256**: 64-character hex string (256 bits) - Good for general use, balance of speed and security
- **sha512**: 128-character hex string (512 bits) - Larger hash, slower but more secure
- **sha1**: 40-character hex string (160 bits) - Legacy, not recommended for security-critical use
- **md5**: 32-character hex string (128 bits) - Legacy, cryptographically broken, don't use for security

## Performance

- SHA256 and SHA512 are fast, suitable for large files
- All algorithms produce consistent output for the same input (deterministic)
- Different from `hash_password()` which uses bcrypt and includes a salt

## Common Use Cases

- **File integrity**: Verify files haven't been modified
- **Deduplication**: Identify identical content
- **Digital signatures**: Hash data before signing
- **Checksum verification**: Compare data integrity
- **Content addressing**: Use hash as unique identifier

## Differences from hash_password()

- `hash()` is deterministic (same input = same output)
- `hash_password()` uses bcrypt with random salt for security
- Use `hash()` for integrity checking, use `hash_password()` for passwords

## See Also

- [hash_password() - Securely hash passwords with bcrypt](/docs/reference/hash_password.md)
- [verify_password() - Verify password against hash](/docs/reference/verify_password.md)
- [encode_base64() - Encode to base64](/docs/reference/encode_base64.md)
