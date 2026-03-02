# verify_password()

Verify a password against a bcrypt hash. Uses constant-time comparison to prevent timing attacks.

## Signature

```duso
verify_password(password, hash)
```

## Parameters

- `password` (string) - The password to verify
- `hash` (string) - The bcrypt hash to compare against (created with `hash_password()`)

## Returns

- `true` if password matches the hash
- `false` if password does not match
- Never throws on mismatch (safe for user login flows)

## Examples

Basic password verification:

```duso
stored_hash = "$2a$10$T7Wdsal81LTgiT0icaSpROpn5nSjapdgSoQTjSp17u9sPUjsB.LPa"
user_input = "userPassword123"

if verify_password(user_input, stored_hash) then
  print("Login successful!")
else
  print("Invalid password")
end
```

User login flow:

```duso
email = input("Email: ")
password = input("Password: ")

// Fetch user from database
user = db.users.find_by_email(email)

if user and verify_password(password, user.password_hash) then
  print("Login successful! Welcome {{user.name}}")
  // Create session/token
else
  print("Invalid email or password")
end
```

Using named arguments:

```duso
correct = verify_password(password = "test123", hash = stored_hash)
```

Safe error handling:

```duso
password = input("Password: ")
hash = load_hash_from_db()

if verify_password(password, hash) then
  // Grant access
  authenticate_user()
else
  // Don't reveal if user exists or password is wrong
  print("Invalid email or password")
  sleep(1)  // Rate limit
end
```

## Security Notes

- Uses constant-time comparison (prevents timing attacks)
- Returns false on mismatch - never throws
- Safe for use in user-facing login flows
- Always pair with rate limiting on login attempts
- Never log or display password values
- Always use HTTPS for password transmission

## See Also

- [hash_password() - Hash a password](/docs/reference/hash_password.md)
- [env() - Read environment variables](/docs/reference/env.md)
