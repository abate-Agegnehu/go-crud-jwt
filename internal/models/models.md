# Models - `internal/models/user.go`

This file defines **what data your application works with**. It is the blueprint for your database table and API requests/responses.

---

## 1. The `User` Struct (Database Table)

```go
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Username  string         `gorm:"unique;not null;size:50" json:"username"`
    Email     string         `gorm:"unique;not null;size:100" json:"email"`
    Password  string         `gorm:"not null" json:"-"`
    FirstName string         `gorm:"size:50" json:"first_name"`
    LastName  string         `gorm:"size:50" json:"last_name"`
    Role      string         `gorm:"default:user;size:20" json:"role"`
    IsActive  bool           `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### What each field does:

| Field        | Go Type         | DB Column    | Purpose                                      |
|-------------|-----------------|--------------|----------------------------------------------|
| `ID`        | `uint`          | `id`         | Auto-increment primary key (1, 2, 3...)      |
| `Username`  | `string`        | `username`   | Unique username, max 50 chars                |
| `Email`     | `string`        | `email`      | Unique email, max 100 chars                  |
| `Password`  | `string`        | `password`   | Hashed password (never sent to API clients)  |
| `FirstName` | `string`        | `first_name` | Optional first name, max 50 chars            |
| `LastName`  | `string`        | `last_name`  | Optional last name, max 50 chars             |
| `Role`      | `string`        | `role`       | `"user"` by default, can be `"admin"`        |
| `IsActive`  | `bool`          | `is_active`  | `true` by default, can deactivate accounts   |
| `CreatedAt` | `time.Time`     | `created_at` | Auto-set when user is created                |
| `UpdatedAt` | `time.Time`     | `updated_at` | Auto-updated when user is modified           |
| `DeletedAt` | `gorm.DeletedAt`| `deleted_at` | Soft delete marker (user isn't physically deleted) |

### The Two Tag Systems:

**`gorm:"..."` tags** - Tell GORM (your ORM) how to create the database table:
- `primaryKey` = This is the primary key
- `unique` = No duplicate values allowed
- `not null` = Cannot be empty
- `size:50` = VARCHAR(50) in database
- `default:user` = Default value is "user"
- `index` = Create a database index for fast lookups

**`json:"..."` tags** - Control JSON output when you send data to the API:
- `json:"id"` = When converted to JSON, use key "id"
- `json:"-"` = NEVER include this field in JSON output (used for Password and DeletedAt)

---

## 2. Request/Response DTOs (Data Transfer Objects)

These are **NOT database tables**. They are temporary structs used only for API input/output.

### `CreateUserRequest` - What the client sends when registering:

```go
type CreateUserRequest struct {
    Username  string `json:"username" binding:"required,min=3,max=50"`
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=6"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}
```

**`binding:"..."` tags** - Gin's automatic validation:
| Tag               | Meaning                              |
|-------------------|--------------------------------------|
| `required`        | Field must be provided               |
| `min=3`           | String must be at least 3 characters |
| `max=50`          | String must be at most 50 characters |
| `email`           | Must be a valid email format         |
| `min=6`           | Password must be at least 6 chars    |
| `omitempty`       | Field is optional (used in Update)   |

**Example JSON the client sends:**
```json
{
    "username": "john123",
    "email": "john@example.com",
    "password": "secret123",
    "first_name": "John",
    "last_name": "Doe"
}
```

### `UpdateUserRequest` - What the client sends when updating:

```go
type UpdateUserRequest struct {
    Username  string `json:"username" binding:"omitempty,min=3,max=50"`
    Email     string `json:"email" binding:"omitempty,email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    IsActive  *bool  `json:"is_active"`
}
```

Key differences from CreateUser:
- `binding:"omitempty"` = All fields are optional (you can update just one field)
- `IsActive *bool` = Pointer to bool so we can distinguish between "not provided" and "set to false"

### `LoginRequest` - What the client sends when logging in:

```go
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}
```

### `LoginResponse` - What the server sends back after login:

```go
type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}
```

**Example JSON response:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "id": 1,
        "username": "john123",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "role": "user",
        "is_active": true,
        "created_at": "2026-08-18T22:00:00Z",
        "updated_at": "2026-08-18T22:00:00Z"
    }
}
```

Notice: `password` is NOT in the response because of `json:"-"` tag.

---

## 3. GORM Hooks (BeforeCreate / BeforeUpdate)

### `BeforeCreate` - Runs automatically BEFORE saving a new user:

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.Password != "" {
        hashedPassword, err := utils.HashPassword(u.Password)
        if err != nil {
            return err
        }
        u.Password = hashedPassword
    }
    return nil
}
```

**Flow:**
1. Client sends password as plain text: `"secret123"`
2. GORM calls `BeforeCreate` automatically
3. `BeforeCreate` hashes it: `$2a$10$N9qo8uLOickgx2ZMRZoMye...` (bcrypt hash)
4. Hashed password is stored in the database
5. Plain text password is never saved

### `BeforeUpdate` - Runs automatically BEFORE updating a user:

```go
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    if tx.Statement.Changed("Password") && u.Password != "" {
        hashedPassword, err := utils.HashPassword(u.Password)
        if err != nil {
            return err
        }
        u.Password = hashedPassword
    }
    return nil
}
```

**Key difference:** It checks `tx.Statement.Changed("Password")` to only hash if the password was actually changed. This prevents re-hashing an already-hashed password.

---

## 4. How This File Connects to Others

```
user_handler.go  --->  receives CreateUserRequest
       |
user_service.go  --->  uses CreateUserRequest to create User
       |
user_repository.go --> saves User to database
       |
database.go       --> creates "users" table from User struct
```

The `User` struct is used EVERYWHERE - it's the core data type of your entire application.
