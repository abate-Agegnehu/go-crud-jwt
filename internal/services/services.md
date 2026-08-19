# Services - `internal/services/user_service.go`

This file contains **all business logic**. It sits between handlers (HTTP layer) and repositories (database layer). It makes decisions, validates rules, and orchestrates operations.

---

## 1. The `UserService` Struct

```go
type UserService struct {
    repo *repositories.UserRepository
}

func NewUserService() *UserService {
    return &UserService{
        repo: repositories.NewUserRepository(),
    }
}
```

The service holds a reference to the repository. It never touches the database directly - always through the repository.

---

## 2. Every Service Method Explained

### `CreateUser` - Register a new user

```go
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
    // Check if email already exists
    exists, err := s.repo.ExistsByEmail(req.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("email already registered")
    }

    // Check if username already exists
    exists, err = s.repo.ExistsByUsername(req.Username)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("username already taken")
    }

    // Create user
    user := &models.User{
        Username:  req.Username,
        Email:     req.Email,
        Password:  req.Password, // Will be hashed by BeforeCreate hook
        FirstName: req.FirstName,
        LastName:  req.LastName,
    }

    err = s.repo.Create(user)
    if err != nil {
        return nil, err
    }

    return user, nil
}
```

**Business rules enforced:**
1. Email must be unique across all users
2. Username must be unique across all users
3. Password is passed as plain text - the `BeforeCreate` hook in the model hashes it

**Flow:**
```
Handler calls service.CreateUser(req)
    |
    v
Check email uniqueness  --> 400 "email already registered"
Check username uniqueness --> 400 "username already taken"
    |
    v
Build User struct from request
    |
    v
repo.Create(user)  --> BeforeCreate hook hashes password
    |
    v
Return created user (with ID, timestamps, etc.)
```

---

### `Login` - Authenticate and get JWT token

```go
func (s *UserService) Login(req *models.LoginRequest) (string, *models.User, error) {
    // Find user by email
    user, err := s.repo.FindByEmail(req.Email)
    if err != nil {
        return "", nil, err
    }
    if user == nil {
        return "", nil, errors.New("invalid email or password")
    }

    // Check if user is active
    if !user.IsActive {
        return "", nil, errors.New("account is deactivated")
    }

    // Verify password
    if !utils.CheckPasswordHash(req.Password, user.Password) {
        return "", nil, errors.New("invalid email or password")
    }

    // Generate JWT token
    token, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
    if err != nil {
        return "", nil, err
    }

    return token, user, nil
}
```

**Security measures:**
1. Uses the SAME error message for wrong email AND wrong password ("invalid email or password"). This prevents attackers from knowing whether an email exists.
2. Checks `IsActive` - deactivated accounts can't log in.
3. Uses `utils.CheckPasswordHash()` which uses bcrypt to compare.

**Flow:**
```
Handler calls service.Login(req)
    |
    v
FindByEmail(email)  --> nil? Return "invalid email or password"
    |
    v
Check IsActive      --> false? Return "account is deactivated"
    |
    v
CheckPasswordHash   --> false? Return "invalid email or password"
    |
    v
GenerateToken       --> Create JWT with user info
    |
    v
Return (token, user, nil)
```

---

### `GetUserByID` - Find a single user

```go
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not found")
    }
    return user, nil
}
```

Simple wrapper - converts `nil` user into a "user not found" error.

---

### `GetUsers` - Paginated list

```go
func (s *UserService) GetUsers(page, limit int) ([]models.User, int64, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }
    return s.repo.FindAll(page, limit)
}
```

**Input validation:** Ensures page >= 1 and limit is between 1-100 (default 10).

---

### `UpdateUser` - Update user fields

```go
func (s *UserService) UpdateUser(id uint, req *models.UpdateUserRequest) (*models.User, error) {
    // Check if user exists
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, errors.New("user not found")
    }

    // Build updates map
    updates := make(map[string]interface{})

    if req.Username != "" {
        existingUser, err := s.repo.FindByUsername(req.Username)
        if err != nil {
            return nil, err
        }
        if existingUser != nil && existingUser.ID != id {
            return nil, errors.New("username already taken")
        }
        updates["username"] = req.Username
    }

    if req.Email != "" {
        existingUser, err := s.repo.FindByEmail(req.Email)
        if err != nil {
            return nil, err
        }
        if existingUser != nil && existingUser.ID != id {
            return nil, errors.New("email already registered")
        }
        updates["email"] = req.Email
    }

    if req.FirstName != "" {
        updates["first_name"] = req.FirstName
    }

    if req.LastName != "" {
        updates["last_name"] = req.LastName
    }

    if req.IsActive != nil {
        updates["is_active"] = *req.IsActive
    }

    err = s.repo.Update(id, updates)
    if err != nil {
        return nil, err
    }

    return s.repo.FindByID(id)
}
```

**Business rules:**
1. User must exist
2. If changing username, check it's not taken by ANOTHER user (`existingUser.ID != id`)
3. If changing email, check it's not taken by ANOTHER user
4. Only update fields that were provided (non-empty)
5. Return the UPDATED user (re-fetch from DB)

**Key detail:** `existingUser.ID != id` - This allows a user to keep their own username/email. Without this check, updating your own profile would fail.

---

### `DeleteUser` - Soft delete

```go
func (s *UserService) DeleteUser(id uint) error {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    return s.repo.SoftDelete(id)
}
```

Soft delete: Sets `deleted_at` timestamp. User can be recovered.

### `HardDeleteUser` - Permanent delete

```go
func (s *UserService) HardDeleteUser(id uint) error {
    user, err := s.repo.FindByID(id)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    return s.repo.Delete(id)
}
```

Hard delete: Physically removes the row from the database. **Cannot be undone.**

---

## 3. Layered Architecture Pattern

```
Handler (HTTP)  -->  Service (Business)  -->  Repository (Data)  -->  Database
     |                     |                       |
Parse JSON          Validate rules            Run SQL queries
Return JSON         Make decisions            Return results
                    Call utilities
```

**Why this separation?**
- **Handlers** only deal with HTTP (parsing, status codes, JSON)
- **Services** only deal with business logic (validation, rules, decisions)
- **Repositories** only deal with database queries
- Each layer can be tested independently
- You can swap the database without changing business logic
- You can swap HTTP framework without changing business logic

---

## 4. How This Connects to Others

```
user_handler.go  -->  calls service.CreateUser(req)
                   calls service.Login(req)
                   calls service.GetUserByID(id)
    |
    v
user_service.go  -->  validates business rules
                   calls repo.Create(user)
                   calls repo.FindByEmail(email)
                   calls utils.CheckPasswordHash(...)
                   calls utils.GenerateToken(...)
    |
    v
user_repository.go -->  runs SQL queries
jwt_utils.go       -->  creates/validates JWT tokens
password_utils.go  -->  hashes/checks passwords
```
