package redis

import (
	"MainApp/settings"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr        string        `yaml:"addr"`
	Password    string        `yaml:"password"`
	User        string        `yaml:"user"`
	DB          int           `yaml:"db"`
	MaxRetries  int           `yaml:"max_retries"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	Timeout     time.Duration `yaml:"timeout"`
}

func NewClient(ctx context.Context, cfg Config) (*redis.Client, error) {
	db := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Username:     cfg.User,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := db.Ping(ctx).Err(); err != nil {
		fmt.Printf("failed to connect to redis server: %s\n", err.Error())
		return nil, err
	}

	return db, nil
}

func GetFreeContainer(ctx context.Context, rdb *redis.Client) (string, error) {
	containers := settings.TestContainers
	for _, c := range containers {
		key := fmt.Sprintf("container:%s", c)

		oldStatus, err := rdb.GetSet(ctx, key, "busy").Result()
		if err != nil && err != redis.Nil {
			return "", err
		}

		if oldStatus == "free" || err == redis.Nil {
			return c, nil
		}
	}

	return "", fmt.Errorf("нет свободных контейнеров")
}

func SetContainerStatus(ctx context.Context, rdb *redis.Client, container, status string) error {
	key := fmt.Sprintf("container:%s", container)
	if status == "free" {
		return rdb.Set(ctx, key, "free", 0).Err()
	} else if status == "busy" {
		return rdb.Set(ctx, key, "busy", 0).Err()
	} else {
		return rdb.Set(ctx, key, "busy", 0).Err()
	}
}
