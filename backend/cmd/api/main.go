// Package main is the entrypoint for the CPM API server.
//
// @title           CPM API
// @version         0.1
// @description     Backend API for the Construction Project Manager application.
// @BasePath        /
package main

import (
	"log"

	"github.com/Priyatna-repository/cpm-go-react/backend/internal/config"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/database"
	"github.com/Priyatna-repository/cpm-go-react/backend/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	routes.Setup(r)

	log.Printf("starting server on :%s (env=%s)", cfg.AppPort, cfg.AppEnv)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
