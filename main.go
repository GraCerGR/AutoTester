package main

import "fmt"

func main() {
	ctx, redisClient, err := Runner()
	if err != nil {
		fmt.Errorf("При запуске runner'а произошла ошибка: %s", err)
		return
	}

	//Прослушивание очереди

	//Скачивание решения и вариантов сайта

	Executor(ctx, redisClient)
}