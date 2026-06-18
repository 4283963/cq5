package database

import (
	"incubator-backend/internal/config"
	applogger "incubator-backend/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	cfg := config.GlobalConfig.Database

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		applogger.Errorf("connect database failed: %v", err)
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		applogger.Errorf("get sql db failed: %v", err)
		return err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	DB = db
	applogger.Info("database connected successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
