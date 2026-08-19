# Middleware - `internal/middleware/auth_middleware.go`

Middleware runs **BEFORE your handler** for every request in a route group. Think of it as a bouncer checking IDs at the door.

---

## 1. `AuthMiddleware()` - JWT Authentication

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... check token ...
        c.Next()
    }
}
```

### Step-by-step flow:

**Step 1: Get the Authorization header**
```go
authHeader := c.GetHeader("Authorization")
if authHeader == "" {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
    c.Abort()
    return
}
```
- Every protected request must include: `Authorization: Bearer <token>`
- `c.Abort()` stops the request - the handler never runs

**Step 2: Parse "Bearer" format**
```go
parts := strings.Split(authHeader, " ")
if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use Bearer token"})
    c.Abort()
    return
}
token := parts[1]
```

Valid format: `Bearer eyJhbGciOiJIUzI1NiIs...`
- Split by space: `["Bearer", "eyJhbGciOiJIUzI1NiIs..."]`
- First part must be "Bearer" (case-insensitive)
- Second part is the token

**Step 3: Validate the JWT token**
```go
claims, err := utils.ValidateToken(token)
if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
    c.Abort()
    return
}
```
- Calls `utils.ValidateToken()` which checks if the token is signed correctly and not expired
- Returns the decoded claims (user ID, username, email, role)

**Step 4: Store user info in context**
```go
c.Set("user_id", claims.UserID)
c.Set("username", claims.Username)
c.Set("email", claims.Email)
c.Set("role", claims.Role)
c.Next()
```
- `c.Set()` stores data that handlers can read later with `c.Get("user_id")`
- `c.Next()` passes the request to the next handler

---

## 2. `RoleMiddleware()` - Role-Based Access Control

```go
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
            c.Abort()
            return
        }

        roleStr := role.(string)
        allowed := false
        for _, r := range allowedRoles {
            if r == roleStr {
                allowed = true
                break
            }
        }

        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Usage in main.go:**
```go
admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
```

**Flow:**
1. Reads the `role` that `AuthMiddleware` stored in context
2. Checks if the role is in the allowed list
3. If not allowed, returns 403 Forbidden

**Example:** `RoleMiddleware("admin")` only allows users with `role = "admin"`.

---

## 3. `c.Abort()` vs `c.Next()`

| Method     | What it does                                      |
|-----------|--------------------------------------------------|
| `c.Next()` | Pass request to the next handler in the chain    |
| `c.Abort()`| Stop processing - handler never runs             |

**Request flow with middleware:**
```
Request arrives
    |
    v
AuthMiddleware
    |-- c.Abort() --> Returns 401, request stops here
    |-- c.Next()  --> Continues to...
    |
    v
RoleMiddleware
    |-- c.Abort() --> Returns 403, request stops here
    |-- c.Next()  --> Continues to...
    |
    v
Handler (e.g., GetUsers)
```

---

## 4. Middleware Chaining in main.go

```go
protected := router.Group("/api")
protected.Use(middleware.AuthMiddleware())
{
    protected.GET("/users", userHandler.GetUsers)
}
```

This means:
1. ANY request to `/api/*` goes through `AuthMiddleware` first
2. If token is valid, the actual handler runs
3. If token is invalid, 401 is returned

```go
admin := router.Group("/api/admin")
admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
```

Here, TWO middlewares run in sequence:
1. `AuthMiddleware` checks the JWT token
2. `RoleMiddleware` checks if the user is an admin

---

## 5. How This Connects to Others

```
HTTP Request with "Authorization: Bearer eyJ..."
    |
    v
auth_middleware.go  -->  validates token
    |                    calls utils.ValidateToken()
    |                    stores user_id, role in context
    v
user_handler.go    -->  reads c.Get("user_id") and c.Get("role")
    |
    v
jwt_utils.go       -->  decodes the JWT token
```

The middleware is the bridge between the HTTP world and your application logic. It protects routes so only authenticated (and authorized) users can access them.
