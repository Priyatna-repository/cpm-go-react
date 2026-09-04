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

	r.Static("/uploads", cfg.UploadDir)
	r.GET("/health", handlers.HealthCheck)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authService := services.NewAuthService(db, cfg)
	authHandler := handlers.NewAuthHandler(authService, cfg)

	permService := services.NewPermissionService(db)
	permHandler := handlers.NewPermissionHandler(permService)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		lookupService := services.NewLookupService(db)
		lookupHandler := handlers.NewLookupHandler(lookupService)

		lookups := api.Group("/lookups")
		lookups.Use(middleware.RequireAuth(authService))
		{
			lookups.GET("/countries", lookupHandler.ListCountries)
			lookups.GET("/currencies", lookupHandler.ListCurrencies)
		}

		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleLogin)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		uploadService := services.NewUploadService(cfg)
		ownerCompanyService := services.NewOwnerCompanyService(db, uploadService)
		ownerCompanyHandler := handlers.NewOwnerCompanyHandler(ownerCompanyService)

		ownerCompany := api.Group("/owner-company")
		ownerCompany.Use(middleware.RequireAuth(authService))
		{
			ownerCompany.GET("", middleware.RequirePermission(permService, "owner_company.view"), ownerCompanyHandler.Get)
			ownerCompany.PUT("", middleware.RequirePermission(permService, "owner_company.edit"), ownerCompanyHandler.Update)
		}

		userLookupService := services.NewUserLookupService(db)
		clientCompanyService := services.NewClientCompanyService(db, uploadService)
		clientCompanyHandler := handlers.NewClientCompanyHandler(clientCompanyService, userLookupService)

		clientCompanies := api.Group("/client-companies")
		clientCompanies.Use(middleware.RequireAuth(authService))
		{
			clientCompanies.GET("", middleware.RequirePermission(permService, "client_company.view"), clientCompanyHandler.List)
			clientCompanies.GET("/:id", middleware.RequirePermission(permService, "client_company.view"), clientCompanyHandler.Get)
			clientCompanies.POST("", middleware.RequirePermission(permService, "client_company.create"), clientCompanyHandler.Create)
			clientCompanies.PUT("/:id", middleware.RequirePermission(permService, "client_company.edit"), clientCompanyHandler.Update)
			clientCompanies.DELETE("/:id", middleware.RequirePermission(permService, "client_company.archive"), clientCompanyHandler.Archive)
			clientCompanies.POST("/:id/restore", middleware.RequirePermission(permService, "client_company.restore"), clientCompanyHandler.Restore)
			clientCompanies.DELETE("/:id/force", middleware.RequirePermission(permService, "client_company.delete"), clientCompanyHandler.ForceDelete)
		}

		lookups.GET("/client-users", middleware.RequirePermission(permService, "client_company.view"), clientCompanyHandler.ListClientUsers)

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
