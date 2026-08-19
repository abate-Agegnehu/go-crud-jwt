# Main - `cmd/main/main.go`

This is the **entry point** of your application. When you run `go run cmd/main/main.go`, this is what executes.

---

## 1. `main()` Function

```go
func main() {
    // Load configuration
    config.LoadConfig()

    // Initialize database
    database.InitDB()

    // Create router
    router := gin.Default()

    // Set up routes
    setupRoutes(router)

    // Start server
    log.Printf("Server starting on port %s", config.AppConfig.Port)
    if err := router.Run(":" + config.AppConfig.Port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

**Execution order (critical to understand):**

```
1. config.LoadConfig()    -->  Reads .env, sets config.AppConfig
2. database.InitDB()      -->  Connects to PostgreSQL, creates tables
3. gin.Default()          -->  Creates HTTP router with logging & recovery
4. setupRoutes(router)    -->  Registers all API routes
5. router.Run(":8080")    -->  Starts HTTP server, listens for requests
```

**If any step fails:**
- `LoadConfig()` fails silently (warns, uses defaults)
- `InitDB()` calls `log.Fatal` (crashes if DB is down)
- `router.Run()` calls `log.Fatal` (crashes if port is in use)

---

## 2. `setupRoutes()` Function

```go
func setupRoutes(router *gin.Engine) {
    userHandler := handlers.NewUserHandler()

    // Public routes
    auth := router.Group("/api/auth")
    {
        auth.POST("/register", userHandler.Register)
        auth.POST("/login", userHandler.Login)
    }

    // Protected routes
    protected := router.Group("/api")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.GET("/users/me", userHandler.GetCurrentUser)
        protected.GET("/users", userHandler.GetUsers)
        protected.GET("/users/:id", userHandler.GetUser)
        protected.PUT("/users/:id", userHandler.UpdateUser)
        protected.DELETE("/users/:id", userHandler.DeleteUser)
    }

    // Admin only routes
    admin := router.Group("/api/admin")
    admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
    {
        admin.GET("/users", userHandler.GetUsers)
        admin.DELETE("/users/:id", userHandler.DeleteUser)
    }

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "message": "Server is running",
        })
    })
}
```

---

## 3. Route Groups Explained

### Public Routes (`/api/auth`)
```go
auth := router.Group("/api/auth")
auth.POST("/register", userHandler.Register)
auth.POST("/login", userHandler.Login)
```
No middleware - anyone can call these.

### Protected Routes (`/api`)
```go
protected := router.Group("/api")
protected.Use(middleware.AuthMiddleware())
```
JWT middleware required - must have valid token.

### Admin Routes (`/api/admin`)
```go
admin := router.Group("/api/admin")
admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
```
JWT + Admin role required.

---

## 4. Complete API Routes

| Method | Path              | Auth Required | Role     | Handler                |
|--------|-------------------|---------------|----------|------------------------|
| POST   | `/api/auth/register` | No        | Any      | `Register`             |
| POST   | `/api/auth/login`    | No        | Any      | `Login`                |
| GET    | `/api/users/me`      | Yes       | Any      | `GetCurrentUser`       |
| GET    | `/api/users`         | Yes       | Any      | `GetUsers`             |
| GET    | `/api/users/:id`     | Yes       | Any      | `GetUser`              |
| PUT    | `/api/users/:id`     | Yes       | Own/Admin| `UpdateUser`           |
| DELETE | `/api/users/:id`     | Yes       | Own/Admin| `DeleteUser`           |
| GET    | `/api/admin/users`   | Yes       | Admin    | `GetUsers`             |
| DELETE | `/api/admin/users/:id`| Yes      | Admin    | `DeleteUser`           |
| GET    | `/health`            | No        | Any      | Inline (200 OK)        |

---

## 5. `gin.Default()` vs `gin.New()`

| Function        | Includes                                   |
|----------------|--------------------------------------------|
| `gin.Default()` | Logger (request logging) + Recovery (panic recovery) |
| `gin.New()`     | Nothing - bare router                      |

Always use `gin.Default()` unless you want to customize middleware.

---

## 6. Request Flow Example

```
Client sends: GET /api/users/5
Headers: Authorization: Bearer eyJhbGci...

    |
    v
gin.Router matches: /api/users/:id
    |
    v
middleware.AuthMiddleware()
    - Extracts token from header
    - Validates JWT
    - Sets user_id, role in context
    |
    v
handlers.UserHandler.GetUser(c)
    - Reads :id from URL ("5")
    - Calls service.GetUserByID(5)
    - Returns user as JSON
    |
    v
Client receives: 200 OK
{
    "id": 5,
    "username": "john123",
    ...
}
```

---

## 7. How This Connects to Everything

```
main.go
    |
    +-- config.LoadConfig()
    |       |
    |       v
    |   config.go  -->  reads .env
    |
    +-- database.InitDB()
    |       |
    |       v
    |   database.go  -->  connects to PostgreSQL
    |
    +-- setupRoutes()
    |       |
    |       v
    |   handlers/  -->  HTTP request handlers
    |   middleware/ -->  Auth checks
    |
    +-- router.Run()
            |
            v
        gin framework  -->  listens for HTTP requests
```

main.go is the conductor - it initializes everything and starts the server.
