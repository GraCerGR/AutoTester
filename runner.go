package main

import (
	"MainApp/conteinermanager"
	dockercompose "MainApp/dockercompose"
	myredis "MainApp/redis"
	"MainApp/settings"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func Runner() (context.Context, *redis.Client, error) {

	// ---- Runner ----
	ctx := context.Background()
	//Запуск компоуза с гридом, реддисом и контейнеров с сайтами
	if err := dockercompose.StartCompose(ctx, "./dockercompose"); err != nil {
		fmt.Printf("Ошибка запуска Selenium Grid: %v\n", err)
		return nil, nil, err
	}

	//Сборка образов для тестовых контейнеров
	for _, stack := range settings.Stacks {
		if err := conteinermanager.DockerBuild(ctx, settings.ChooseImageTag(stack), settings.ChooseImageFilePath(stack)+settings.ChooseImageFile(stack), "."); err != nil {
			fmt.Printf("Ошибка запуска Selenium Grid: %v\n", err)
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

	//Запуск worker очереди
	StartQueueWorker(ctx, redisClient)

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
