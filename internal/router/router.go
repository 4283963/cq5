package router

import (
	"incubator-backend/internal/cache"
	"incubator-backend/internal/handler"
	"incubator-backend/internal/middleware"
	"incubator-backend/internal/repository"
	"incubator-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	r.Use(middleware.Logger())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.Recovery())

	userRepo := repository.NewUserRepository()
	deviceRepo := repository.NewDeviceRepository()
	configRepo := repository.NewDeviceConfigRepository()
	reportRepo := repository.NewDeviceReportRepository()
	alarmRepo := repository.NewAlarmRepository()

	statusCache := cache.NewDeviceStatusCacheManager()
	configCache := cache.NewDeviceConfigCacheManager()

	authService := service.NewAuthService(userRepo)
	alarmService := service.NewAlarmService(alarmRepo)
	safetyService := service.NewSafetyService(alarmRepo, deviceRepo)
	deviceService := service.NewDeviceService(deviceRepo)
	configService := service.NewDeviceConfigService(configRepo, deviceRepo, configCache)
	reportService := service.NewDeviceReportService(reportRepo, deviceRepo, statusCache, safetyService)

	authHandler := handler.NewAuthHandler(authService)
	alarmHandler := handler.NewAlarmHandler(alarmService, safetyService)
	deviceHandler := handler.NewDeviceHandler(deviceService)
	configHandler := handler.NewDeviceConfigHandler(configService)
	reportHandler := handler.NewDeviceReportHandler(reportService)

	authMiddleware := middleware.NewJWTAuthMiddleware(authService)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMiddleware.AuthRequired(), authHandler.GetCurrentUser)
		}

		devices := api.Group("/devices")
		devices.Use(authMiddleware.AuthRequired())
		{
			devices.GET("", deviceHandler.List)
			devices.GET("/:id", deviceHandler.GetByID)
			devices.POST("", authMiddleware.AdminRequired(), deviceHandler.Create)
			devices.PUT("/:id", authMiddleware.AdminRequired(), deviceHandler.Update)
			devices.DELETE("/:id", authMiddleware.AdminRequired(), deviceHandler.Delete)

			devices.POST("/:device_id/recover", authMiddleware.AdminRequired(), alarmHandler.RecoverDevice)
		}

		configs := api.Group("/device-configs")
		configs.Use(authMiddleware.AuthRequired())
		{
			configs.GET("", configHandler.List)
			configs.GET("/:id", configHandler.GetByID)
			configs.GET("/device/:device_id", configHandler.GetByDeviceID)
			configs.GET("/code/:device_code", configHandler.GetByDeviceCode)
			configs.POST("", authMiddleware.AdminRequired(), configHandler.Create)
			configs.PUT("/:id", authMiddleware.AdminRequired(), configHandler.Update)
			configs.DELETE("/:id", authMiddleware.AdminRequired(), configHandler.Delete)

			configs.GET("/cache/:device_code", configHandler.GetCachedConfig)
			configs.POST("/cache/sync/:device_code", authMiddleware.AdminRequired(), configHandler.SyncToCache)
		}

		reports := api.Group("/device-reports")
		reports.Use(authMiddleware.AuthRequired())
		{
			reports.GET("/device/:device_id", reportHandler.ListByDevice)
			reports.GET("/latest/code/:device_code", reportHandler.GetLatest)
			reports.GET("/cache/:device_code", reportHandler.GetCachedStatus)
			reports.POST("/sync-all", authMiddleware.AdminRequired(), reportHandler.SyncAllFromDB)
		}

		alarms := api.Group("/alarms")
		alarms.Use(authMiddleware.AuthRequired())
		{
			alarms.GET("", alarmHandler.List)
			alarms.GET("/:id", alarmHandler.GetByID)
			alarms.GET("/active/device/:device_id", alarmHandler.GetActiveByDevice)
			alarms.POST("/:id/resolve", authMiddleware.AdminRequired(), alarmHandler.Resolve)
		}
	}

	deviceAPI := r.Group("/api/device")
	{
		deviceAPI.POST("/report/:device_code", reportHandler.Report)
	}
}
