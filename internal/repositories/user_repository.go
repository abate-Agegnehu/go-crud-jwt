// internal/repositories/user_repository.go
package repositories

import (
    "errors"

    "go-crud-jwt/internal/database"
    "go-crud-jwt/internal/models"

    "gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository() *UserRepository {
    return &UserRepository{db: database.DB}
}

// CRUD Operations
func (r *UserRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
    var user models.User
    err := r.db.Where("username = ?", username).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

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

func (r *UserRepository) Update(id uint, updates map[string]interface{}) error {
    return r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepository) Delete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) SoftDelete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
    var count int64
    err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
    return count > 0, err
}

func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
    var count int64
    err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
    return count > 0, err
}