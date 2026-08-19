# Go Module - `go.mod`

This file defines your **project's identity and dependencies**. It's like `package.json` in Node.js or `requirements.txt` in Python.

---

## 1. Module Declaration

```go
module go-crud-jwt
```

This is your project's import path. When other files import packages, they use this prefix:

```go
import "go-crud-jwt/internal/models"
//         ^^^^^^^^^^
//         this matches the module name
```

---

## 2. Go Version

```go
go 1.21
```

Minimum Go version required. Your code uses features from Go 1.21.

---

## 3. Direct Dependencies

```go
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/joho/godotenv v1.5.1
    golang.org/x/crypto v0.18.0
    gorm.io/driver/postgres v1.5.4
    gorm.io/gorm v1.25.5
)
```

| Package              | Version | What it does                         |
|---------------------|---------|--------------------------------------|
| `gin-gonic/gin`     | v1.9.1  | HTTP web framework (router, handlers)|
| `golang-jwt/jwt/v5` | v5.2.0  | Create and validate JWT tokens       |
| `joho/godotenv`     | v1.5.1  | Load `.env` files                    |
| `golang.org/x/crypto`| v0.18.0| Bcrypt password hashing              |
| `gorm.io/driver/postgres`| v1.5.4| PostgreSQL driver for GORM        |
| `gorm.io/gorm`      | v1.25.5 | ORM (Object-Relational Mapper)       |

These are packages YOUR code directly imports.

---

## 4. Indirect Dependencies

```go
require (
    github.com/bytedance/sonic v1.9.1 // indirect
    github.com/gin-contrib/sse v0.1.0 // indirect
    github.com/go-playground/validator/v10 v10.14.0 // indirect
    // ... many more
)
```

These are packages that YOUR dependencies need. You don't use them directly, but they're required for your dependencies to work.

**Example:** `gin` uses `validator` for input validation, so `validator` is an indirect dependency.

---

## 5. How Dependencies Are Managed

```bash
# Add a new dependency
go get github.com/some/package

# Download all dependencies
go mod download

# Clean up unused dependencies and add missing ones
go mod tidy

# See why a dependency is needed
go mod why github.com/some/package
```

---

## 6. `go.sum` File

The `go.sum` file (not shown) contains cryptographic checksums for every dependency. It ensures:
- Dependencies haven't been tampered with
- Everyone on your team downloads the same versions
- Builds are reproducible

---

## 7. Project Structure with Module

```
go-crud-jwt/                    <-- Module root (go.mod is here)
    go.mod
    go.sum
    .env
    cmd/
        main/
            main.go             <-- imports "go-crud-jwt/internal/config"
    internal/
        config/
            config.go           <-- package config
        database/
            database.go         <-- package database
        handlers/
            user_handler.go     <-- package handlers
        middleware/
            auth_middleware.go   <-- package middleware
        models/
            user.go             <-- package models
        repositories/
            user_repository.go  <-- package repositories
        services/
            user_service.go     <-- package services
        utils/
            jwt_utils.go        <-- package utils
            password_utils.go   <-- package utils
```

Every package imports others using the module prefix:
```go
// In handlers/user_handler.go
import "go-crud-jwt/internal/models"
import "go-crud-jwt/internal/services"
```
