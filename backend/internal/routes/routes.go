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
		// client company endpoints
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
		lookups.GET("/internal-users", middleware.RequirePermission(permService, "project.manage_access"), clientCompanyHandler.ListInternalUsers)

		labelService := services.NewLabelService(db)
		labelHandler := handlers.NewLabelHandler(labelService)
		// labels endpoints
		labels := api.Group("/labels")
		labels.Use(middleware.RequireAuth(authService))
		{
			labels.GET("", middleware.RequirePermission(permService, "labels.view"), labelHandler.List)
			labels.GET("/:id", middleware.RequirePermission(permService, "labels.view"), labelHandler.Get)
			labels.POST("", middleware.RequirePermission(permService, "labels.create"), labelHandler.Create)
			labels.PUT("/:id", middleware.RequirePermission(permService, "labels.edit"), labelHandler.Update)
			labels.DELETE("/:id", middleware.RequirePermission(permService, "labels.archive"), labelHandler.Archive)
			labels.POST("/:id/restore", middleware.RequirePermission(permService, "labels.restore"), labelHandler.Restore)
			labels.DELETE("/:id/force", middleware.RequirePermission(permService, "labels.delete"), labelHandler.ForceDelete)
		}

		projectService := services.NewProjectService(db)
		projectHandler := handlers.NewProjectHandler(projectService, permService)
		// project endpoints
		projects := api.Group("/projects")
		projects.Use(middleware.RequireAuth(authService))
		{
			projects.GET("", middleware.RequirePermission(permService, "project.view"), projectHandler.List)
			projects.GET("/:id",
				middleware.RequirePermission(permService, "project.view"),
				middleware.RequireProjectAccess(projectService),
				projectHandler.Get,
			)
			projects.POST("", middleware.RequirePermission(permService, "project.create"), projectHandler.Create)
			projects.PUT("/:id",
				middleware.RequirePermission(permService, "project.edit"),
				middleware.RequireProjectAccess(projectService),
				projectHandler.Update,
			)
			projects.DELETE("/:id",
				middleware.RequirePermission(permService, "project.archive"),
				middleware.RequireProjectAccess(projectService),
				projectHandler.Archive,
			)
			projects.POST("/:id/restore",
				middleware.RequirePermission(permService, "project.restore"),
				middleware.RequireProjectAccess(projectService),
				projectHandler.Restore,
			)
			projects.DELETE("/:id/force",
				middleware.RequirePermission(permService, "project.delete"),
				middleware.RequireProjectAccess(projectService),
				projectHandler.ForceDelete,
			)
			projects.PUT("/:id/access",
				middleware.RequirePermission(permService, "project.manage_access"),
				middleware.RequireProjectAccess(projectService),
				projectHandler.UpdateAccess,
			)
		}

		userService := services.NewUserService(db, uploadService)
		// internal user endpoints
		internalUserHandler := handlers.NewUserHandler(userService, services.UserScopeInternal)
		users := api.Group("/users")
		users.Use(middleware.RequireAuth(authService))
		{
			users.GET("", middleware.RequirePermission(permService, "user.view"), internalUserHandler.List)
			users.GET("/:id", middleware.RequirePermission(permService, "user.view"), internalUserHandler.Get)
			users.POST("", middleware.RequirePermission(permService, "user.create"), internalUserHandler.Create)
			users.PUT("/:id", middleware.RequirePermission(permService, "user.edit"), internalUserHandler.Update)
			users.DELETE("/:id", middleware.RequirePermission(permService, "user.archive"), internalUserHandler.Archive)
			users.POST("/:id/restore", middleware.RequirePermission(permService, "user.restore"), internalUserHandler.Restore)
			users.DELETE("/:id/force", middleware.RequirePermission(permService, "user.delete"), internalUserHandler.ForceDelete)
		}

		clientUserHandler := handlers.NewUserHandler(userService, services.UserScopeClient)
		clientUsers := api.Group("/client-user-accounts")
		clientUsers.Use(middleware.RequireAuth(authService))
		{
			clientUsers.GET("", middleware.RequirePermission(permService, "client_user.view"), clientUserHandler.List)
			clientUsers.GET("/:id", middleware.RequirePermission(permService, "client_user.view"), clientUserHandler.Get)
			clientUsers.POST("", middleware.RequirePermission(permService, "client_user.create"), clientUserHandler.Create)
			clientUsers.PUT("/:id", middleware.RequirePermission(permService, "client_user.edit"), clientUserHandler.Update)
			clientUsers.DELETE("/:id", middleware.RequirePermission(permService, "client_user.archive"), clientUserHandler.Archive)
			clientUsers.POST("/:id/restore", middleware.RequirePermission(permService, "client_user.restore"), clientUserHandler.Restore)
			clientUsers.DELETE("/:id/force", middleware.RequirePermission(permService, "client_user.delete"), clientUserHandler.ForceDelete)
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
