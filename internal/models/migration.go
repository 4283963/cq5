package models

import (
	"incubator-backend/pkg/database"
	"incubator-backend/pkg/logger"
)

func AutoMigrate() error {
	logger.Info("starting database migration...")

	err := database.DB.AutoMigrate(
		&User{},
		&Device{},
		&DeviceConfig{},
		&DeviceReport{},
	)
	if err != nil {
		logger.Errorf("migration failed: %v", err)
		return err
	}

	logger.Info("database migration completed successfully")
	return nil
}
