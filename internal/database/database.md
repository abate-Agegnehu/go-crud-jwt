# Database - `internal/database/database.go`

This file **connects to PostgreSQL** and creates/migrates your database tables.

---

## 1. The Global DB Variable

```go
var DB *gorm.DB
```

A global database connection that every other file uses to talk to PostgreSQL. When `repositories` need to query the database, they use `database.DB`.

---

## 2. `InitDB()` Function - Step by Step

### Step 1: Build the connection string (DSN)

```go
cfg := config.AppConfig
dsn := fmt.Sprintf(
    "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
    cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
)
```

**DSN** (Data Source Name) is the connection string PostgreSQL needs.

**Example DSN:**
```
host=localhost user=postgres password=mypass dbname=go_crud_db port=5432 sslmode=disable TimeZone=UTC
```

| Part          | Meaning                                      |
|--------------|----------------------------------------------|
| `host`       | Where PostgreSQL is running                  |
| `user`       | Database username                            |
| `password`   | Database password                            |
| `dbname`     | Which database to use                        |
| `port`       | PostgreSQL port (default 5432)               |
| `sslmode=disable` | No SSL (fine for local dev, use `require` in production) |
| `TimeZone=UTC` | Use UTC timezone for timestamps            |

### Step 2: Open the connection

```go
DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

- `postgres.Open(dsn)` = PostgreSQL driver for GORM
- `logger.Info` = Log all SQL queries to console (great for debugging)
- Change to `logger.Silent` in production to hide query logs

### Step 3: Auto-migrate (create/update tables)

```go
err = DB.AutoMigrate(&models.User{})
```

This reads the `User` struct and:
- If table doesn't exist -> creates it
- If table exists but is missing columns -> adds them
- If table exists and matches -> does nothing

**It does NOT:**
- Remove columns that were deleted from the struct
- Change column types
- Drop the table

For those, you'd use manual migrations.

### Step 4: Fatal on error

```go
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}
```

`log.Fatal` prints the error and **crashes the program**. If the database is down, there's no point continuing.

---

## 3. What GORM Does Under the Hood

When `AutoMigrate(&models.User{})` runs, GORM translates the `User` struct into SQL:

```sql
CREATE TABLE IF NOT EXISTS "users" (
    "id" BIGSERIAL PRIMARY KEY,
    "username" VARCHAR(50) NOT NULL UNIQUE,
    "email" VARCHAR(100) NOT NULL UNIQUE,
    "password" VARCHAR NOT NULL,
    "first_name" VARCHAR(50),
    "last_name" VARCHAR(50),
    "role" VARCHAR(20) DEFAULT 'user',
    "is_active" BOOLEAN DEFAULT true,
    "created_at" TIMESTAMPTZ,
    "updated_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users"("deleted_at");
```

---

## 4. How This Connects to Others

```
config.go
    |
    v
database.go (InitDB)  <-- called once in main.go
    |
    v
database.DB (global)   <-- used by repositories
    |
    v
user_repository.go  -->  r.db.Create(user)
                       r.db.Find(&user, id)
                       r.db.Where("email = ?", email)
```

The database connection is established ONCE at startup, stored in `database.DB`, and shared with all repository files.
