package main

import (
	"MainApp/checker"
	"fmt"
	"os"
)

func main() {

	stop, profiles := cmdCommands()
	if stop {
		return
	}

	_, _, err := Runner(profiles) // ctx, redisClient,
	if err != nil {
		fmt.Printf("При запуске runner'а произошла ошибка: %s", err)
		return
	}
}

func cmdCommands() (bool, []string) {
	if len(os.Args) > 1 && os.Args[1] == "up" {

		noRedis := false
		noKafka := false
		noSelenium := false

		for _, arg := range os.Args[2:] {
			switch arg {
			case "--no-redis":
				noRedis = true
			case "--no-kafka":
				noKafka = true
			case "--no-selenium":
				noSelenium = true
			}
		}

		var profiles []string

		if !noRedis {
			profiles = append(profiles, "redis")
		}
		if !noKafka {
			profiles = append(profiles, "kafka")
		}
		if !noSelenium {
			profiles = append(profiles, "selenium")
		}

		return false, profiles
	}

	if len(os.Args) > 1 && os.Args[1] == "parsexml" {

		if len(os.Args) < 3 {
			fmt.Println("Ошибка. Для прасинга: go run . parsexml <папка> <index_bool>")
			return true, nil
		}

		if len(os.Args) == 3 {
			if err := checker.ParseFolder(os.Args[2]); err != nil {
				fmt.Println("Ошибка:", err)
				os.Exit(1)
			}
		} else {
			if err := checker.ParseFolder(os.Args[2], os.Args[3]); err != nil {
				fmt.Println("Ошибка:", err)
				os.Exit(1)
			}
		}

		return true, nil
	}

	if len(os.Args) > 1 {
		fmt.Println("Доступные команды:\nup\nparsexml")
		return true, nil
	}

	return false, []string{"redis", "kafka", "selenium"}

}
