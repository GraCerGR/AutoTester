package main

import (
	"MainApp/checker"
	"MainApp/classes"
	redis "MainApp/messagebrokers/redis"
	"context"
	"fmt"
	"os"
)

func main() {

	if cmdCommands() == true {
		return
	}

	ctx, redisClient, err := Runner()
	if err != nil {
		fmt.Printf("При запуске runner'а произошла ошибка: %s", err)
		return
	}

	//Прослушивание очереди kafka
	go func() {
		err := StartAttemptKafkaListener(ctx,
			func(ctx context.Context, a classes.Attempt) error {
				return redis.EnqueueAttempt(ctx, redisClient, a)
			},
		)

		if err != nil && ctx.Err() == nil {
			fmt.Printf("Прослушивание закончено с ошибкой: %v", err)
		}
	}()

	<-ctx.Done()
}

func cmdCommands() bool {
	if len(os.Args) > 1 && os.Args[1] == "parsexml" {

		if len(os.Args) < 3 {
			fmt.Println("Ошибка. Для прасинга выполните команду: go run . parsexml <папка_с_xml> <index_bool>")
			return true
		}
		if len(os.Args) == 3 {
			if err := checker.ParseFolder(os.Args[2]); err != nil {
				fmt.Println("Ошибка при обработке папки:", err)
				os.Exit(1)
			}
		} else if len(os.Args) == 4 {
			if err := checker.ParseFolder(os.Args[2], os.Args[3]); err != nil {
				fmt.Println("Ошибка при обработке папки:", err)
				os.Exit(1)
			}
		}
		fmt.Println(len(os.Args))
		return true

	} else if len(os.Args) > 1 {

		fmt.Println("Неверная команда. Доступные команды:\nparsexml")
		return true

	} else {
		return false
	}
}
