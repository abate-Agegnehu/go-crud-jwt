// cmd/main/main.go
package main

import (
    "log"

    "go-crud-jwt/internal/config"
    "go-crud-jwt/internal/database"
    "go-crud-jwt/internal/handlers"
    "go-crud-jwt/internal/middleware"

    "github.com/gin-gonic/gin"
)

func main() {
    // Load configuration
    config.LoadConfig()

    // Initialize database
    database.InitDB()

    // Create router
    router := gin.Default()

    // Set up routes
    setupRoutes(router)

    // Start server
    log.Printf("Server starting on port %s", config.AppConfig.Port)
    if err := router.Run(":" + config.AppConfig.Port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}

func setupRoutes(router *gin.Engine) {
    userHandler := handlers.NewUserHandler()

    // Public routes
    auth := router.Group("/api/auth")
    {
        auth.POST("/register", userHandler.Register)
        auth.POST("/login", userHandler.Login)
    }

    // Protected routes (authentication required)
    protected := router.Group("/api")
    protected.Use(middleware.AuthMiddleware())
    {
        // User routes
        protected.GET("/users/me", userHandler.GetCurrentUser)
        protected.GET("/users", userHandler.GetUsers)
        protected.GET("/users/:id", userHandler.GetUser)
        protected.PUT("/users/:id", userHandler.UpdateUser)
        protected.DELETE("/users/:id", userHandler.DeleteUser)
    }

    // Admin only routes
    admin := router.Group("/api/admin")
    admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
    {
        // Add admin-specific routes here
        admin.GET("/users", userHandler.GetUsers)
        admin.DELETE("/users/:id", userHandler.DeleteUser)
    }

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "message": "Server is running",
        })
    })
}