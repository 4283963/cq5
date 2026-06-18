package redis

import (
	"context"
	"incubator-backend/internal/config"
	"incubator-backend/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() error {
	cfg := config.GlobalConfig.Redis

	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		logger.Errorf("connect redis failed: %v", err)
		return err
	}

	logger.Info("redis connected successfully")
	return nil
}

func GetClient() *redis.Client {
	return Client
}
