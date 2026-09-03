package routes

import (
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Priyatna-repository/cpm-go-react/backend/docs"
)

func Setup(r *gin.Engine) {
	r.GET("/health", handlers.HealthCheck)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		_ = api // module route groups (auth, projects, tasks, ...) are registered here as each module ships
	}
}
