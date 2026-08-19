# Config - `internal/config/config.go`

This file **loads all settings** your application needs from environment variables or a `.env` file.

---

## 1. The `Config` Struct

```go
type Config struct {
    DBHost     string
    DBUser     string
    DBPassword string
    DBName     string
    DBPort     string
    JWTSecret  string
    JWTExpiry  time.Duration
    Port       string
}
```

| Field        | Purpose                                  | Example Value          |
|-------------|------------------------------------------|------------------------|
| `DBHost`    | PostgreSQL server address                | `localhost`            |
| `DBUser`    | Database username                        | `postgres`             |
| `DBPassword`| Database password                        | `mypassword`           |
| `DBName`    | Database name                            | `go_crud_db`           |
| `DBPort`    | PostgreSQL port                          | `5432`                 |
| `JWTSecret` | Secret key to sign JWT tokens            | `my-super-secret-key`  |
| `JWTExpiry` | How long a token stays valid             | `24h` (24 hours)       |
| `Port`      | HTTP server port                         | `8080`                 |

---

## 2. The Global Variable

```go
var AppConfig *Config
```

This is a **global pointer** to the config. Once `LoadConfig()` runs, every other file in your app can access settings via `config.AppConfig.JWTSecret`, `config.AppConfig.DBHost`, etc.

Why a pointer (`*Config`)? Because it starts as `nil` and gets assigned only after `LoadConfig()` runs.

---

## 3. `LoadConfig()` Function

```go
func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: .env file not found, using environment variables")
    }
    // ...
}
```

**Step 1:** Try to load `.env` file using `godotenv`. If not found, just warn and continue (uses system environment variables instead).

```go
    jwtExpiry, err := time.ParseDuration(os.Getenv("JWT_EXPIRY"))
    if err != nil {
        jwtExpiry = 24 * time.Hour
    }
```

**Step 2:** Parse `JWT_EXPIRY` from string to `time.Duration`. If not set or invalid, default to 24 hours.

```go
    AppConfig = &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBUser:     getEnv("DB_USER", "postgres"),
        // ...
    }
```

**Step 3:** Build the Config struct using `getEnv()` for each field.

---

## 4. `getEnv()` Helper Function

```go
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

This reads an environment variable. If it's not set or empty, returns the default value.

**Example:**
- If `DB_HOST` is set to `"192.168.1.100"` -> returns `"192.168.1.100"`
- If `DB_HOST` is not set -> returns `"localhost"` (the default)

---

## 5. The `.env` File

Your `.env` file (in project root) should look like:

```
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=go_crud_db
DB_PORT=5432
JWT_SECRET=my-super-secret-key-that-is-long
JWT_EXPIRY=24h
PORT=8080
```

**Important:** Never commit `.env` to git. Add it to `.gitignore`.

---

## 6. How This Connects to Other Files

```
.env file
    |
    v
config.go  (LoadConfig)  -->  sets config.AppConfig
    |
    v
database.go  (InitDB)    -->  reads config.AppConfig.DBHost, etc. to connect to DB
main.go      (main)      -->  reads config.AppConfig.Port to start server
jwt_utils.go             -->  reads config.AppConfig.JWTSecret and JWTExpiry
```

The config is loaded ONCE at startup in `main.go`, then every other package reads from the global `AppConfig`.
