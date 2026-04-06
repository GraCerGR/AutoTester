package dockercompose

import (
	"MainApp/conteinermanager"
	"MainApp/settings"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/segmentio/kafka-go"
)

func StartCompose(ctx context.Context, composeDir string) error {
	absDir, err := filepath.Abs(composeDir)
	if err != nil {
		return fmt.Errorf("Не удалось получить абсолютный путь: %w", err)
	}

	fmt.Println("Запуск контейнеров docker compose")

	if err := conteinermanager.RunCmd(ctx, "docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "up", "-d", "--scale", "selenium-node-chrome="+settings.SeleniumNodeChromeNumber); err != nil { //, "--scale", "selenium-node-chrome=5"
		return fmt.Errorf("docker compose up failed: %w", err)
	}

	// Ждём Kafka
	fmt.Println("Ожидание готовности Kafka")
	if err := WaitForKafka(30*time.Second, settings.KafkaBrokers[0]); err != nil {
		fmt.Println("Kafka не стал готова вовремя:")
		_ = conteinermanager.RunCmd(ctx, "docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "logs", "kafka")
		return fmt.Errorf("Kafka не готов: %w", err)
	}

	fmt.Println("Kafka готов")

	fmt.Println("Ожидание готовности Selenium Hub")
	if err := WaitForHub(settings.HubWaitTimeout); err != nil {
		fmt.Println("Hub не стал готов вовремя:")
		_ = conteinermanager.RunCmd(ctx, "docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "logs")
		return fmt.Errorf("Hub не готов: %w", err)
	}

	fmt.Println("Selenium Grid запущен и готов по адресу:", settings.HubURL)
	return nil
}

func StopCompose(ctx context.Context, composeDir string) error {
	absDir, err := filepath.Abs(composeDir)
	if err != nil {
		return fmt.Errorf("не удалось получить абсолютный путь: %w", err)
	}

	fmt.Println("=== Остановка Selenium Grid ===")
	if err := conteinermanager.RunCmd(ctx, "docker", "compose", "-f", filepath.Join(absDir, "docker-compose.yml"), "down", "-v"); err != nil {
		return fmt.Errorf("docker compose down failed: %w", err)
	}

	fmt.Println("Selenium Grid остановлен")
	return nil
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
