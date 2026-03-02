# hash_password()

Hash a password using bcrypt for secure storage.

## Signature

```duso
hash_password(password [, cost])
```

## Parameters

- `password` (string) - The password to hash
- `cost` (optional, number) - Computational cost factor. Range: 4-31, default: 10
  - Lower values (4-6) are faster but less secure
  - Higher values (12-14) are slower but more secure against brute force
  - Each increment roughly doubles the time to compute

## Returns

Bcrypt hash string (always different due to random salt, even for identical passwords)

## Examples

Basic password hashing:

```duso
password = "userPassword123"
hash = hash_password(password)
print(hash)  // $2a$10$T7Wdsal81LTgiT0icaSpROpn5nSjapdgSoQTjSp17u9sPUjsB.LPa
```

Store password hash in database:

```duso
user_password = input("Enter password: ")
hash = hash_password(user_password)

// Store hash in database
db.users.create({
  email = "user@example.com",
  password_hash = hash
})
```

Using named arguments with custom cost:

```duso
// More secure for sensitive applications
secure_hash = hash_password(password = "mySecret", cost = 14)
```

Different cost levels:

```duso
// Fast for development/testing
dev_hash = hash_password("password", cost = 4)

// Balanced
prod_hash = hash_password("password", cost = 10)

// Extra secure for high-value accounts
admin_hash = hash_password("password", cost = 14)
```

## Security Notes

- Uses bcrypt with random salt - each hash is unique
- Cost factor automatically adjusts to modern hardware
- Never store plain passwords - always use hashes
- Use cost 10+ for production systems
- Even with identical passwords, hashes will differ

## See Also

- [verify_password() - Check password against hash](/docs/reference/verify_password.md)
- [encode_base64() - Encode to base64](/docs/reference/encode_base64.md)
