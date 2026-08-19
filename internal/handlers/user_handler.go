// internal/handlers/user_handler.go
package handlers

import (
    "net/http"
    "strconv"

    "go-crud-jwt/internal/models"
    "go-crud-jwt/internal/services"

    "github.com/gin-gonic/gin"
)

type UserHandler struct {
    service *services.UserService
}

func NewUserHandler() *UserHandler {
    return &UserHandler{
        service: services.NewUserService(),
    }
}

// @Summary Register new user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User registration details"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
    var req models.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.CreateUser(&req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, user)
}

// @Summary Login user
// @Tags Users
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    token, user, err := h.service.Login(&req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, models.LoginResponse{
        Token: token,
        User:  *user,
    })
}

// @Summary Get current user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 404 {object} map[string]interface{}
// @Router /api/users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    user, err := h.service.GetUserByID(userID.(uint))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, user)
}

// @Summary Get user by ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 404 {object} map[string]interface{}
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    user, err := h.service.GetUserByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, user)
}

// @Summary Get all users
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    users, total, err := h.service.GetUsers(page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "users": users,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}

// @Summary Update user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body models.UpdateUserRequest true "Update details"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Check if user is updating themselves or is admin
    userID, _ := c.Get("user_id")
    role, _ := c.Get("role")
    if userID.(uint) != uint(id) && role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
        return
    }

    var req models.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.UpdateUser(uint(id), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, user)
}

// @Summary Delete user
// @Tags Users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]interface{}
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    // Check if user is deleting themselves or is admin
    userID, _ := c.Get("user_id")
    role, _ := c.Get("role")
    if userID.(uint) != uint(id) && role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own account"})
        return
    }

    err = h.service.DeleteUser(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}