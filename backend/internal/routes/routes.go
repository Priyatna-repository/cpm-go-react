package routes

import (
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/handlers"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/middleware"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "github.com/Priyatna-repository/cpm-go-react/backend/docs"
)

func Setup(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	r.GET("/health", handlers.HealthCheck)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authService := services.NewAuthService(db, cfg)
	authHandler := handlers.NewAuthHandler(authService, cfg, db)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		api.GET("/me", middleware.RequireAuth(authService), authHandler.Me)
	}
}
