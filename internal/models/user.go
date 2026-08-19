// internal/models/user.go
package models

import (
    "time"

    "go-crud-jwt/internal/utils"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Username  string         `gorm:"unique;not null;size:50" json:"username"`
    Email     string         `gorm:"unique;not null;size:100" json:"email"`
    Password  string         `gorm:"not null" json:"-"` // "-" hides from JSON
    FirstName string         `gorm:"size:50" json:"first_name"`
    LastName  string         `gorm:"size:50" json:"last_name"`
    Role      string         `gorm:"default:user;size:20" json:"role"`
    IsActive  bool           `gorm:"default:true" json:"is_active"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Request/Response DTOs
type CreateUserRequest struct {
    Username  string `json:"username" binding:"required,min=3,max=50"`
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=6"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

type UpdateUserRequest struct {
    Username  string `json:"username" binding:"omitempty,min=3,max=50"`
    Email     string `json:"email" binding:"omitempty,email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    IsActive  *bool  `json:"is_active"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

// BeforeCreate hook to hash password
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

// BeforeUpdate hook to hash password if changed
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