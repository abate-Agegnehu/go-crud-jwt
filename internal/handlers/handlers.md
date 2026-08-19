# Handlers - `internal/handlers/user_handler.go`

This file **handles HTTP requests** - it receives API calls, parses input, calls the service layer, and sends back JSON responses. Think of it as the "waiter" that takes orders and brings back food.

---

## 1. The `UserHandler` Struct

```go
type UserHandler struct {
    service *services.UserService
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        service: services.NewUserService(),
    }
}
```

The handler holds a reference to `UserService`. When a request comes in, the handler calls the service to do the actual work.

---

## 2. Every Handler Method Explained

### `Register` - POST `/api/auth/register`

```go
func (h *UserHandler) Register(c *gin.Context) {
    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.CreateUser(&req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, user)
}
```

**Flow:**
1. `c.ShouldBindJSON(&req)` = Read the JSON body and fill the `CreateUserRequest` struct. If validation fails (missing email, short password, etc.), return 400 error.
2. `h.service.CreateUser(&req)` = Tell the service to create the user. If email/username already taken, return error.
3. `c.JSON(http.StatusCreated, user)` = Return 201 with the created user as JSON.

**Request:**
```json
POST /api/auth/register
{
    "username": "john123",
    "email": "john@example.com",
    "password": "secret123"
}
```

**Success Response (201):**
```json
{
    "id": 1,
    "username": "john123",
    "email": "john@example.com",
    "role": "user",
    "is_active": true
}
```

---

### `Login` - POST `/api/auth/login`

```go
func (h *UserHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    token, user, err := h.service.Login(&req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, models.LoginResponse{
        Token: token,
        User:  *user,
    })
}
```

**Flow:**
1. Parse email + password from JSON body
2. Service verifies credentials and generates JWT token
3. Return token + user info

**Success Response (200):**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": { "id": 1, "username": "john123", ... }
}
```

---

### `GetCurrentUser` - GET `/api/users/me`

```go
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    user, err := h.service.GetUserByID(userID.(uint))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, user)
}
```

**Flow:**
1. `c.Get("user_id")` = Read the user ID that the **auth middleware** stored in the context
2. `userID.(uint)` = Type assertion - convert `interface{}` to `uint`
3. Fetch and return the user

---

### `GetUser` - GET `/api/users/:id`

```go
func (h *UserHandler) GetUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    // ...
    user, err := h.service.GetUserByID(uint(id))
    // ...
    c.JSON(http.StatusOK, user)
}
```

**Flow:**
1. `c.Param("id")` = Get the `:id` from the URL (e.g., `/api/users/5` -> `"5"`)
2. `strconv.ParseUint` = Convert string `"5"` to number `5`
3. Fetch and return the user

---

### `GetUsers` - GET `/api/users?page=1&limit=10`

```go
func (h *UserHandler) GetUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    users, total, err := h.service.GetUsers(page, limit)
    // ...
    c.JSON(http.StatusOK, gin.H{
        "users": users,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}
```

**Flow:**
1. Read query parameters with defaults (page=1, limit=10 if not provided)
2. Fetch paginated users from service
3. Return users + metadata

**Response (200):**
```json
{
    "users": [{ "id": 1, ... }, { "id": 2, ... }],
    "total": 50,
    "page": 1,
    "limit": 10
}
```

---

### `UpdateUser` - PUT `/api/users/:id`

```go
func (h *UserHandler) UpdateUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    // ...

    // Authorization check
    userID, _ := c.Get("user_id")
    role, _ := c.Get("role")
    if userID.(uint) != uint(id) && role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
        return
    }

    var req models.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.UpdateUser(uint(id), &req)
    // ...
}
```

**Authorization logic:** You can only update YOUR OWN profile unless you're an admin.

---

### `DeleteUser` - DELETE `/api/users/:id`

```go
func (h *UserHandler) DeleteUser(c *gin.Context) {
    // Same authorization check as UpdateUser
    // Soft-deletes the user
    c.JSON(http.StatusNoContent, nil)
}
```

Returns `204 No Content` on success (no body, just status code).

---

## 3. HTTP Status Codes Used

| Code   | Constant              | When                                    |
|--------|----------------------|------------------------------------------|
| `200`  | `StatusOK`           | Success                                  |
| `201`  | `StatusCreated`      | Resource created (register)              |
| `204`  | `StatusNoContent`    | Success with no body (delete)            |
| `400`  | `BadRequest`         | Bad input / validation failed            |
| `401`  | `Unauthorized`       | Not logged in / bad credentials          |
| `403`  | `Forbidden`          | Logged in but not allowed                |
| `404`  | `NotFound`           | Resource doesn't exist                   |
| `500`  | `InternalServerError`| Server error                             |

---

## 4. `gin.H` Shortcut

`gin.H` is just `map[string]interface{}`:
```go
gin.H{"error": "not found"}
// is equivalent to
map[string]interface{}{"error": "not found"}
// or in modern Go:
map[string]any{"error": "not found"}
```

---

## 5. How This Connects to Others

```
HTTP Request (e.g., POST /api/auth/register)
    |
    v
main.go (router)  -->  routes to UserHandler.Register
    |
    v
user_handler.go    -->  parses JSON, calls service
    |
    v
user_service.go    -->  business logic
    |
    v
user_repository.go -->  database queries
```

Handlers should NEVER contain business logic. They only:
1. Parse input
2. Call the service
3. Return the response
