# JWT Utilities - `internal/utils/jwt_utils.go`

This file handles **creating and validating JWT tokens**. JWT (JSON Web Token) is how your API remembers who is logged in.

---

## 1. What is a JWT Token?

A JWT token has 3 parts separated by dots:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImpvaG4ifQ.signature
   --------------------------  -----------------------------------  ------------------
              HEADER                     PAYLOAD                    SIGNATURE
```

- **Header:** Algorithm used (`HS256`)
- **Payload:** Your data (user ID, username, email, role, expiry time)
- **Signature:** Ensures the token wasn't tampered with

---

## 2. The `Claims` Struct

```go
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

**Custom fields:** `UserID`, `Username`, `Email`, `Role` - the data you want to store in the token.

**`jwt.RegisteredClaims`** adds standard fields:
- `ExpiresAt` - When the token expires
- `IssuedAt` - When the token was created
- `Issuer` - Who created the token

**Example decoded payload:**
```json
{
    "user_id": 1,
    "username": "john123",
    "email": "john@example.com",
    "role": "user",
    "exp": 1692432000,
    "iat": 1692345600,
    "iss": "go-crud-api"
}
```

---

## 3. `GenerateToken()` - Create a JWT

```go
func GenerateToken(userID uint, username, email, role string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Email:    email,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AppConfig.JWTExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "go-crud-api",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.AppConfig.JWTSecret))
}
```

**Step by step:**
1. Build `Claims` with user data + expiry time
2. `jwt.SigningMethodHS256` = HMAC-SHA256 (symmetric key signing)
3. `token.SignedString([]byte(secret))` = Sign the token with your secret key
4. Returns the full token string

**What the token contains:**
```json
{
    "user_id": 1,
    "username": "john123",
    "email": "john@example.com",
    "role": "user",
    "exp": 1692432000,
    "iat": 1692345600,
    "iss": "go-crud-api"
}
```

---

## 4. `ValidateToken()` - Verify a JWT

```go
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(config.AppConfig.JWTSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

**Step by step:**
1. `jwt.ParseWithClaims` = Parse the token and extract claims
2. The callback function verifies the signing method is HMAC (security check)
3. Returns the secret key so the library can verify the signature
4. Check `token.Valid` - false if expired or tampered with
5. Return the decoded claims

**Fails when:**
- Token is expired (`exp` has passed)
- Token was tampered with (signature doesn't match)
- Token was signed with a different secret
- Token is malformed

---

## 5. How JWT Auth Works End-to-End

```
LOGIN:
    User sends email + password
        |
    Service verifies credentials
        |
    GenerateToken() creates JWT
        |
    JWT sent back to client

PROTECTED REQUEST:
    Client sends: Authorization: Bearer eyJhbGci...
        |
    AuthMiddleware extracts token
        |
    ValidateToken(token) decodes it
        |
    Claims contain user_id, role, etc.
        |
    Stored in context for handlers to use
```

---

## 6. Security Notes

- **HS256** = Same key signs AND verifies (symmetric). Good for single-server apps.
- **RS256** = Public/private key pair (asymmetric). Better for distributed systems.
- The secret (`JWT_SECRET`) must be kept private. If leaked, anyone can forge tokens.
- Tokens are stateless - no database lookup needed to verify them.
- Tokens have an expiry - after that, user must log in again.

---

## 7. How This Connects to Others

```
user_service.go (Login)
    |
    v
jwt_utils.go (GenerateToken)  -->  creates JWT on login
    |
    v
auth_middleware.go (AuthMiddleware)
    |
    v
jwt_utils.go (ValidateToken)  -->  validates JWT on every request
    |
    v
user_handler.go  -->  reads claims from context
```
