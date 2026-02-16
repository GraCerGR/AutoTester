package redis

import (
	"MainApp/classes"
	"MainApp/settings"
	"context"
	"encoding/json"
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

func GetFreeContainer(ctx context.Context, rdb *redis.Client, containers []settings.Container, stack string) (string, error) {

	lua := `
		local stack = redis.call('HGET', KEYS[1], 'stack')
		local status = redis.call('HGET', KEYS[1], 'status')
		if (ARGV[1] == '' or stack == ARGV[1]) and status == 'free' then
		redis.call('HSET', KEYS[1], 'status', 'busy')
		return 1
		end
		return 0`

	for _, c := range containers {
		key := fmt.Sprintf("container:%s", c.Name)
		res, err := rdb.Eval(ctx, lua, []string{key}, stack).Result()
		if err != nil {
			res, err = rdb.Eval(ctx, lua, []string{key}, stack).Result()
			if err != nil {
				return "", err
			}
		}
		if num, ok := res.(int64); ok && num == 1 {
			return c.Name, nil
		}
	}

	return "", fmt.Errorf("нет свободных контейнеров для стека%q", stack)
}

func SetContainerStatus(ctx context.Context, rdb *redis.Client, container, status string) error {
	key := fmt.Sprintf("container:%s", container)
	switch status {
	case "free":
		return rdb.HSet(ctx, key, "status", status).Err()
	case "busy":
		return rdb.HSet(ctx, key, "status", status).Err()
	default:
		return rdb.HSet(ctx, key, "status", "busy").Err()
	}
}

// Очередь контейнеров и их статусов
func InitOrUpdateContainer(ctx context.Context, rdb *redis.Client, c settings.Container, status string) error {
	key := fmt.Sprintf("container:%s", c.Name)
	return rdb.HSet(ctx, key, map[string]interface{}{
		"stack":  c.Stack,
		"status": status,
	}).Err()
}

// Очередь попыток на проверку
func EnqueueAttempt(ctx context.Context, rdb *redis.Client, a classes.Attempt) error {
	stack := a.ProgrammingLanguageName
	key := fmt.Sprintf("queue:%s", stack)

	data, err := json.Marshal(a)
	if err != nil {
		return fmt.Errorf("marshal attempt: %w", err)
	}

	return rdb.RPush(ctx, key, data).Err()
}
