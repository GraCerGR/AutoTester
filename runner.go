package main

import (
	myredis "MainApp/redis"
	dockercompose "MainApp/dockercompose"
	"MainApp/settings"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func Runner() (context.Context, *redis.Client, error) {

	// ---- Runner ----

	//Запуск компоуза с гридом, тестом и сайтом
	//Запускается при запуске всего хаба
	if err := dockercompose.StartCompose("./dockercompose"); err != nil {
		fmt.Printf("Ошибка запуска Selenium Grid: %v\n", err)
		return nil, nil, err
	}

	//Редис
	ctx := context.Background()
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