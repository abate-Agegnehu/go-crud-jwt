# Repositories - `internal/repositories/user_repository.go`

This file handles **all database operations** for the User table. It's the only file that directly talks to the database through GORM.

---

## 1. The `UserRepository` Struct

```go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository() *UserRepository {
    return &UserRepository{db: database.DB}
}
```

- Holds a reference to the global database connection
- Created via constructor `NewUserRepository()`

---

## 2. Every CRUD Operation Explained

### `Create` - INSERT a new user

```go
func (r *UserRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}
```

**Generated SQL:**
```sql
INSERT INTO "users" ("username","email","password","first_name","last_name","role","is_active","created_at","updated_at")
VALUES ('john123', 'john@example.com', '$2a$10$...', 'John', 'Doe', 'user', true, NOW(), NOW())
RETURNING "id";
```

Note: The `BeforeCreate` hook runs automatically, hashing the password before INSERT.

---

### `FindByID` - SELECT one user by ID

```go
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil  // Not found = return (nil, nil)
    }
    return &user, err
}
```

**Generated SQL:**
```sql
SELECT * FROM "users" WHERE id = 1 AND "deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1;
```

**Key pattern:** When record not found, returns `(nil, nil)` instead of an error. This lets the caller check `if user == nil`.

---

### `FindByEmail` - SELECT one user by email

```go
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}
```

**Generated SQL:**
```sql
SELECT * FROM "users" WHERE email = 'john@example.com' AND "deleted_at" IS NULL LIMIT 1;
```

The `?` is a **placeholder** that prevents SQL injection. GORM safely escapes the value.

---

### `FindByUsername` - SELECT one user by username

```go
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
    var user models.User
    err := r.db.Where("username = ?", username).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}
```

Same pattern as `FindByEmail`.

---

### `FindAll` - SELECT with pagination

```go
func (r *UserRepository) FindAll(page, limit int) ([]models.User, int64, error) {
    var users []models.User
    var total int64

    offset := (page - 1) * limit

    // Count total records
    if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Get paginated records
    err := r.db.Offset(offset).Limit(limit).Order("id desc").Find(&users).Error
    return users, total, err
}
```

**Example with page=2, limit=10:**
- `offset` = (2-1) * 10 = 10 (skip first 10 records)
- `total` = 50 (total users in DB)
- Returns users 11-20

**Generated SQL:**
```sql
SELECT COUNT(*) FROM "users" WHERE "deleted_at" IS NULL;
SELECT * FROM "users" WHERE "deleted_at" IS NULL ORDER BY id DESC LIMIT 10 OFFSET 10;
```

---

### `Update` - UPDATE specific fields

```go
func (r *UserRepository) Update(id uint, updates map[string]interface{}) error {
    return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}
```

**Called with:**
```go
updates := map[string]interface{}{
    "username": "new_username",
    "email":    "new@email.com",
}
repo.Update(1, updates)
```

**Generated SQL:**
```sql
UPDATE "users" SET "username" = 'new_username', "email" = 'new@email.com' WHERE id = 1;
```

---

### `Delete` - DELETE (hard delete, permanent)

```go
func (r *UserRepository) Delete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}
```

**Generated SQL:**
```sql
DELETE FROM "users" WHERE id = 1;
```

**WARNING:** This permanently removes the record from the database.

---

### `SoftDelete` - DELETE (soft delete, recoverable)

```go
func (r *UserRepository) SoftDelete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}
```

Wait, this looks the same as `Delete`! The difference is in the `User` model - it has `DeletedAt gorm.DeletedAt`. When a model has this field, GORM automatically does soft delete:

**Generated SQL (soft delete):**
```sql
UPDATE "users" SET "deleted_at" = NOW() WHERE id = 1;
```

The record is still there, just marked as deleted. All queries automatically filter: `WHERE "deleted_at" IS NULL`.

---

### `ExistsByEmail` - Check if email exists

```go
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
    var count int64
    err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
    return count > 0, err
}
```

**Generated SQL:**
```sql
SELECT COUNT(*) FROM "users" WHERE email = 'john@example.com' AND "deleted_at" IS NULL;
```

Returns `true` if count > 0 (email already taken).

---

## 3. GORM Query Building Pattern

GORM uses a **fluent/chainable API**:

```go
r.db                    // Start with DB
    .Model(&models.User{})  // Specify the table
    .Where("email = ?", email)  // Add WHERE clause
    .First(&user)              // Execute and get first result
    .Error                     // Get any error
```

Each method returns a `*gorm.DB`, allowing you to chain more methods.

---

## 4. How This Connects to Others

```
user_service.go  -->  calls repo.Create(user)
                   calls repo.FindByEmail(email)
                   calls repo.Update(id, updates)
    |
    v
user_repository.go  -->  translates to SQL using GORM
    |
    v
database.go  -->  DB *gorm.DB connection
    |
    v
PostgreSQL  -->  actual database
```

The repository is the ONLY layer that directly uses `database.DB`. Everything else goes through the repository.
