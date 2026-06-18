package main

import (
	"fmt"
	"incubator-backend/internal/config"
	"incubator-backend/internal/models"
	"incubator-backend/internal/repository"
	"incubator-backend/internal/router"
	"incubator-backend/internal/service"
	"incubator-backend/pkg/database"
	"incubator-backend/pkg/logger"
	"incubator-backend/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load("configs/config.yaml"); err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	logger.Init()

	if err := database.Init(); err != nil {
		panic(fmt.Sprintf("init database failed: %v", err))
	}

	if err := redis.Init(); err != nil {
		panic(fmt.Sprintf("init redis failed: %v", err))
	}

	if err := models.AutoMigrate(); err != nil {
		panic(fmt.Sprintf("migrate database failed: %v", err))
	}

	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	if err := authService.InitAdminUser(); err != nil {
		logger.Errorf("init admin user failed: %v", err)
	}

	gin.SetMode(config.GlobalConfig.Server.Mode)
	r := gin.New()
	r.Use(gin.Recovery())

	router.SetupRouter(r)

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	logger.Infof("server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		panic(fmt.Sprintf("start server failed: %v", err))
	}
}
