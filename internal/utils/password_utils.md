# Password Utilities - `internal/utils/password_utils.go`

This file handles **hashing and verifying passwords** using bcrypt.

---

## 1. `HashPassword()` - Hash a plain text password

```go
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}
```

**What it does:**
1. Takes a plain text password (e.g., `"secret123"`)
2. Uses bcrypt to create a one-way hash
3. Returns the hash string

**Example:**
```
Input:    "secret123"
Output:   "$2a$10$N9qo8uLOickgx2ZMRZoMye..." (60 characters)
```

**`bcrypt.DefaultCost` = 10** - This is the "work factor". Higher = slower = more secure against brute force. 10 is the default and takes about 100ms on modern hardware.

**Important:** Hashing is ONE-WAY. You cannot reverse a hash to get the original password.

---

## 2. `CheckPasswordHash()` - Verify a password against a hash

```go
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

**What it does:**
1. Takes the plain text password and the stored hash
2. bcrypt re-hashes the password with the same salt from the stored hash
3. Compares the results
4. Returns `true` if they match, `false` otherwise

**Example usage:**
```go
// During registration
hashedPassword, _ := utils.HashPassword("secret123")
// Store hashedPassword in database

// During login
isValid := utils.CheckPasswordHash("secret123", hashedPassword)  // true
isValid := utils.CheckPasswordHash("wrongpassword", hashedPassword)  // false
```

---

## 3. How Bcrypt Works

```
Plain text: "secret123"
     |
     v
Bcrypt Algorithm:
  1. Generate random salt (e.g., "$2a$10$N9qo8uLOickgx2ZMRZoMye")
  2. Mix salt + password together
  3. Run through key derivation function (thousands of iterations)
  4. Produce hash
     |
     v
Stored in DB: "$2a$10$N9qo8uLOickgx2ZMRZoMye..."
                ^^ ^^
                |  |
                |  +-- Salt (embedded in the hash)
                +----- Algorithm identifier

Verification:
  1. Extract salt from stored hash
  2. Hash the input password with same salt
  3. Compare results
```

**Why bcrypt is secure:**
- Each password gets a unique random salt (prevents rainbow tables)
- Slow by design (prevents brute force)
- No way to reverse the hash

---

## 4. Where Password Hashing Happens

**During registration** (`BeforeCreate` hook in `user.go`):
```
Client sends: "secret123"
    |
BeforeCreate hook: HashPassword("secret123")
    |
Database stores: "$2a$10$N9qo8uLOickgx2ZMRZoMye..."
```

**During login** (`services/user_service.go`):
```
Client sends: "secret123"
    |
service.Login: CheckPasswordHash("secret123", storedHash)
    |
Returns: true/false
```

---

## 5. Common Mistakes to Avoid

| Mistake                           | Why it's bad                              |
|----------------------------------|-------------------------------------------|
| Storing plain text passwords     | If DB is leaked, all passwords exposed    |
| Using MD5/SHA for passwords      | Too fast, vulnerable to brute force       |
| Using same hash for same password| Enables rainbow table attacks             |
| Not using bcrypt's salt feature  | Same password = same hash (predictable)   |

Bcrypt handles all of this automatically - unique salt, slow hashing, no reversal.

---

## 6. How This Connects to Others

```
models/user.go (BeforeCreate hook)
    |
    v
password_utils.go (HashPassword)  -->  hashes on registration
    
user_service.go (Login)
    |
    v
password_utils.go (CheckPasswordHash)  -->  verifies on login
```

This is a utility file - it's stateless, has no dependencies on other project files, and just does one thing well.
