package main

import (
	"MainApp/classes"
	"MainApp/settings"
	"context"
	"fmt"
)

func main() {
	ctx, redisClient, err := Runner()
	if err != nil {
		fmt.Errorf("При запуске runner'а произошла ошибка: %s", err)
		return
	}

	//Прослушивание очереди
	go func() {
		err := StartAttemptInsertListener(ctx, settings.PostgresLink, 
			func(ctx context.Context, a classes.Attempt) error {
			go Executor(ctx, redisClient, a)
			return nil
		},
	)

		if err != nil && ctx.Err() == nil {
			fmt.Errorf("watcher stopped with error: %v", err)
		}
	}()

	<-ctx.Done()
}