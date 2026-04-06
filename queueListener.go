package main

import (
	"MainApp/classes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)


func StartAttemptKafkaListener(ctx context.Context, brokers []string, topic string, groupID string, onNew func(ctx context.Context, a classes.Attempt) error,
) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Printf("Kafka consumer started: topic=%s group=%s", topic, groupID)

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
			log.Printf("bad kafka message at offset=%d: %v", msg.Offset, err)
			_ = r.CommitMessages(ctx, msg)
			continue
		}

		if err := onNew(ctx, a); err != nil {
			log.Printf("processing attempt %s failed: %v", a.Id, err)
			continue
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}
	}
}


func StartAttemptInsertListener(ctx context.Context, pgDSN string, onNew func(ctx context.Context, a classes.Attempt) error) error {
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		return fmt.Errorf("pgxpool.New: %v", err)
	}
	defer pool.Close()

	// Берём отдельное подключение для LISTEN (чтобы уведомления не “пропадали” среди запросов пула)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `LISTEN attempt_insert`)
	if err != nil {
		return fmt.Errorf("LISTEN attempt_insert failed: %v", err)
	}
	log.Println("LISTEN attempt_insert готов")

	for {
		select {
		case <-ctx.Done():
			return err
		default:
		}

		// ждём уведомление (блокирующий вызов)
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			// при разрыве соединения/контекста просто выходим (можно заворачивать в рестарт-луп)
			fmt.Printf("wait notification: %v", err)
			return err
		}
		if notification == nil {
			continue
		}

		id := notification.Payload
		att, err := loadAttemptByID(ctx, pool, id)
		if err != nil {
			log.Printf("load attempt %s failed: %v", id, err)
			continue
		}

		// передаём дальше в Executor
		if err := onNew(ctx, att); err != nil {
			log.Printf("processing attempt %s failed: %v", id, err)
		}
	}
}

func loadAttemptByID(ctx context.Context, pool *pgxpool.Pool, id string) (classes.Attempt, error) {
	var a classes.Attempt

	row := pool.QueryRow(ctx, `
		SELECT
			id,
			created_at,

			solution_git_url,
			solution_git_branch,

			site_git_url,
			site_git_branch,

			variable_with_url,
			programming_language_name,

			timeout_execution,
			timeout_test,

			threads,

			shutdown_condition

		FROM attempt
		WHERE id = $1
	`, id)

	err := row.Scan(
		&a.Id,
		&a.CreatedAt,
		&a.SolutionGit.URL,
		&a.SolutionGit.Branch,
		&a.SiteGit.URL,
		&a.SiteGit.Branch,
		&a.VariableWithURL,
		&a.ProgrammingLanguageName,
		&a.Timeouts.Execution,
		&a.Timeouts.Test,
		&a.Threads,
		&a.ShutdownCondition,
	)
	return a, err
}
