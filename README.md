# go-crud-jwt

A RESTful API built with Go, Gin, GORM, and JWT authentication.

## Features

- User registration and login with JWT
- CRUD operations for users
- Role-based access control (user/admin)
- Password hashing with bcrypt
- PostgreSQL database with GORM

## Project Structure

```
├── cmd/main/           # Application entry point
├── internal/
│   ├── config/         # Configuration loading
│   ├── database/       # Database connection
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/      # Auth & role middleware
│   ├── models/         # Data models & DTOs
│   ├── repositories/   # Database queries
│   ├── services/       # Business logic
│   └── utils/          # Helpers (password hashing, JWT)
```

## API Endpoints

### Public

| Method | Endpoint             | Description       |
|--------|----------------------|-------------------|
| POST   | /api/auth/register   | Register new user |
| POST   | /api/auth/login      | Login user        |
| GET    | /health              | Health check      |

### Protected (requires Bearer token)

| Method | Endpoint             | Description       |
|--------|----------------------|-------------------|
| GET    | /api/users/me        | Get current user  |
| GET    | /api/users           | Get all users     |
| GET    | /api/users/:id       | Get user by ID    |
| PUT    | /api/users/:id       | Update user       |
| DELETE | /api/users/:id       | Delete user       |

## Setup

1. Create a PostgreSQL database
2. Copy `.env.example` to `.env` and configure your database credentials
3. Run the application:

```bash
go run cmd/main/main.go
```

## Environment Variables

| Variable    | Description            | Default                    |
|-------------|------------------------|----------------------------|
| DB_HOST     | Database host          | localhost                  |
| DB_USER     | Database user          | postgres                   |
| DB_PASSWORD | Database password      | -                          |
| DB_NAME     | Database name          | go_crud_db                 |
| DB_PORT     | Database port          | 5432                       |
| JWT_SECRET  | JWT signing secret     | -                          |
| JWT_EXPIRY  | Token expiration       | 24h                        |
| PORT        | Server port            | 8080                       |

## Tech Stack

- **Go 1.21+**
- **Gin** - HTTP framework
- **GORM** - ORM library
- **PostgreSQL** - Database
- **JWT** - Authentication
