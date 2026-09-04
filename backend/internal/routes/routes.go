package routes

import (
	"time"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/handlers"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/middleware"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "github.com/Priyatna-repository/cpm-go-react/backend/docs"
)

func Setup(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", handlers.HealthCheck)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authService := services.NewAuthService(db, cfg)
	authHandler := handlers.NewAuthHandler(authService, cfg)

	permService := services.NewPermissionService(db)
	permHandler := handlers.NewPermissionHandler(permService)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleLogin)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		api.GET("/me", middleware.RequireAuth(authService), authHandler.Me)

		api.GET(
			"/permissions",
			middleware.RequireAuth(authService),
			middleware.RequirePermission(permService, "roles.view"),
			permHandler.ListPermissions,
		)

		roles := api.Group("/roles")
		roles.Use(middleware.RequireAuth(authService))
		{
			roles.GET("", middleware.RequirePermission(permService, "roles.view"), permHandler.ListRoles)
			roles.PUT("/:id/permissions", middleware.RequirePermission(permService, "roles.manage"), permHandler.UpdateRolePermissions)
		}
	}
}
