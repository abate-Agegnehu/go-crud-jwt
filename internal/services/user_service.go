// internal/services/user_service.go
package services

import (
    "errors"

    "go-crud-jwt/internal/models"
    "go-crud-jwt/internal/repositories"
    "go-crud-jwt/internal/utils"
)

type UserService struct {
    repo *repositories.UserRepository
}

func NewUserService() *UserService {
    return &UserService{
        repo: repositories.NewUserRepository(),
    }
}

// Create user with validation
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

// Login user
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

// Get user by ID
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

// Get all users with pagination
func (s *UserService) GetUsers(page, limit int) ([]models.User, int64, error) {
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }
    return s.repo.FindAll(page, limit)
}

// Update user
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
        // Check if username is taken by another user
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
        // Check if email is taken by another user
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

    // Update user
    err = s.repo.Update(id, updates)
    if err != nil {
        return nil, err
    }

    // Fetch updated user
    return s.repo.FindByID(id)
}

// Delete user (soft delete)
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

// Hard delete user (permanent)
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