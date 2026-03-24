package main

import (
	"MainApp/classes"
	myredis "MainApp/redis"
	"MainApp/settings"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Слушает все очереди и запускает Executor, когда контейнер свободен
// Создаёт очереди для каждого стека
func StartQueueWorker(ctx context.Context, rdb *redis.Client) {

	for _, stack := range settings.Stacks {
		go func(stack string) {
			key := fmt.Sprintf("queue:%s", stack)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// ждём элемент из очереди
					data, err := rdb.BLPop(ctx, 1*time.Second, key).Result()
					if err != nil {
						continue
					}
					if len(data) < 2 {
						continue
					}

					var a classes.Attempt
					if err := json.Unmarshal([]byte(data[1]), &a); err != nil {
						fmt.Printf("Ошибка разбора attempt из очереди: %v\n", err)
						continue
					}

					// Поиск свободного тестового контейнера
					containerTests, err := WaitForFreeContainer(ctx, rdb, settings.TestContainers, stack, 1)
					if err != nil {
						fmt.Printf("Не удалось получить свободный контейнер: %v\n", err)
						_ = rdb.LPush(ctx, key, data[1]).Err()
						time.Sleep(1 * time.Second)
						continue
					}

					// Поиск свободного сайтового контейнера
					containerSites, err := WaitForFreeContainer(ctx, rdb, settings.SiteContainers, "", a.Threads)
					if err != nil {
						fmt.Printf("Не удалось получить свободный сайт-контейнер: %v\n", err)
						_ = rdb.LPush(ctx, key, data[1]).Err()
						time.Sleep(1 * time.Second)
						continue
					}

					myredis.SetContainerStatus(ctx, rdb, containerTests[0], "busy")

					for _, c := range containerSites {
						myredis.SetContainerStatus(ctx, rdb, c, "busy")
					}
					go Executor(ctx, rdb, a, containerTests[0], containerSites)
				}
			}
		}(stack)
	}
}

func WaitForFreeContainer(ctx context.Context, rdb *redis.Client, containers []settings.Container, stack string, count int) ([]string, error) {

	var result []string

	for len(result) < count {
		name, err := myredis.GetFreeContainer(ctx, rdb, containers, stack)
		if err == nil {
			result = append(result, name)
		}
		fmt.Printf("Нет свободных контейнеров. Ждём.\n")
		time.Sleep(1 * time.Second)
	}
	return result, nil
}
