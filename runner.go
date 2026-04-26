package main

import (
	"MainApp/containermanager"
	dockercompose "MainApp/dockercompose"
	messagebrokers "MainApp/messagebrokers/kafka"
	myredis "MainApp/messagebrokers/redis"
	"MainApp/settings"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func Runner(profiles []string) (context.Context, *redis.Client, error) {

	// ---- Runner ----
	ctx := context.Background()

	//Запуск компоуза с гридом, реддисом и кафкой
	if len(profiles) != 0 {
		if err := dockercompose.StartCompose(ctx, "./dockercompose", profiles); err != nil {
			fmt.Printf("Ошибка запуска docker compose: %v\n", err)
			return nil, nil, err
		}
	}

	//Сборка образов для тестовых контейнеров
	for _, stack := range settings.Stacks {
		if err := containermanager.DockerBuild(ctx, settings.ChooseImageTag(stack), settings.ChooseImageFilePath(stack)+settings.ChooseImageFile(stack), "."); err != nil {
			fmt.Printf("Ошибка создания образов: %v\n", err)
			return nil, nil, err
		}
	}

	//Редис
	redisClient, err := redisClientStart()
	if err != nil {
		fmt.Printf("Ошибка запуска Redis: %v\n", err)
		return nil, nil, err
	}

	// Пометка в редисе, что все контейнеры свободны
	for _, i := range settings.TestContainers {
		myredis.InitOrUpdateContainer(ctx, redisClient, i, "free")
	}
	for _, i := range settings.SiteContainers {
		myredis.InitOrUpdateContainer(ctx, redisClient, i, "free")
	}

	// Ждём Kafka
	fmt.Println("Ожидание подключения к Kafka")
	if err := WaitForKafka(30*time.Second, settings.KafkaBrokers[0]); err != nil {
		fmt.Println("Kafka не стал готова вовремя:")
		return nil, nil, fmt.Errorf("Kafka не готов: %w", err)
	}
	fmt.Println("Kafka готов")

	fmt.Println("Ожидание подключения к Selenium Hub")
	if err := WaitForHub(settings.HubWaitTimeout); err != nil {
		fmt.Println("Hub не стал готов вовремя:")
		return nil, nil, fmt.Errorf("Hub не готов: %w", err)
	}
	fmt.Println("Selenium Grid запущен и готов по адресу:", settings.HubURL)

	//Запуск очереди
	// Не запускать без kafka
	if dockercompose.Contains(profiles, "kafka") {
		time.Sleep(10 * time.Second)
		if err = messagebrokers.CreateTopicKafka(settings.KafkaBrokers[0], settings.KafkaResultsTopic, settings.KafkaPartitions, settings.KafkaReplicationFactor); err != nil {
			fmt.Printf("Ошибка создания топика Kafka: %v\n", err)
			return nil, nil, err
		}
	}

	StartAttemptKafkaWorker(ctx, redisClient)

	return ctx, redisClient, nil
}

func redisClientStart() (*redis.Client, error) {
	cfg := myredis.Config{
		Addr:        settings.RedisAddr,
		Password:    settings.RedisPassword,
		User:        "",
		DB:          settings.RedisDB,
		MaxRetries:  settings.RedisMaxRetries,
		DialTimeout: settings.DialTimeout,
		Timeout:     settings.RedisTimeout,
	}

	db, err := myredis.NewClient(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to Redis:", cfg.Addr)
	return db, nil
}

func WaitForHub(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(settings.HubStatusURL)
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			return nil
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(settings.HubWaitPollInterval)
	}
	return fmt.Errorf("timeout waiting for hub at %s", settings.HubStatusURL)
}

func WaitForKafka(timeout time.Duration, broker string) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := kafka.Dial("tcp", broker)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		if conn != nil {
			_ = conn.Close()
		}
		time.Sleep(settings.KafkaPollInterval)
	}

	return fmt.Errorf("kafka broker %s не стал готов за %s", broker, timeout)
}
