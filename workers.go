package main

import (
	"MainApp/classes"
	myredis "MainApp/messagebrokers/redis"
	"MainApp/settings"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Слушает все очереди и запускает Executor, когда контейнер свободен
// Создаёт очереди для каждого стека
func StartAttemptKafkaWorker(ctx context.Context, rdb *redis.Client) error {

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  settings.KafkaBrokers,
		Topic:    settings.KafkaAttemptsTopic,
		GroupID:  settings.KafkaGroup,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer r.Close()
	fmt.Printf("Kafka reader started: topic=%s group=%s\n", settings.KafkaAttemptsTopic, settings.KafkaGroup)

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  settings.KafkaBrokers,
		Topic:    settings.KafkaResultsTopic,
		Balancer: &kafka.Hash{},
	})
	defer w.Close()

	fmt.Printf("Kafka writer started: topic=%s\n", settings.KafkaResultsTopic)

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("fetch message: %w", err)
		}

		var a classes.Attempt
		if err := json.Unmarshal(msg.Value, &a); err != nil {
			fmt.Printf("bad kafka message offset=%d: %v", msg.Offset, err)
			_ = r.CommitMessages(ctx, msg)
			continue
		}

		var containerTests []string
		var containerSites []string

		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			containerTests, err = WaitForFreeContainer(ctx, rdb, settings.TestContainers, a.ProgrammingLanguageName, 1)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			containerSites, err = WaitForFreeContainer(ctx, rdb, settings.SiteContainers, "", a.Threads)
			if err != nil {
				_ = myredis.SetContainerStatus(ctx, rdb, containerTests[0], "free")
				time.Sleep(1 * time.Second)
				continue
			}

			break
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			_ = myredis.SetContainerStatus(ctx, rdb, containerTests[0], "free")
			for _, c := range containerSites {
				_ = myredis.SetContainerStatus(ctx, rdb, c, "free")
			}
			return fmt.Errorf("commit message: %w", err)
		}

		myredis.SetContainerStatus(ctx, rdb, containerTests[0], "busy")

		for _, c := range containerSites {
			myredis.SetContainerStatus(ctx, rdb, c, "busy")
		}

		go Executor(ctx, rdb, w, a, containerTests[0], containerSites)
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
