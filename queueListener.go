package main

import (
	"MainApp/classes"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func StartAttemptInsertListener(ctx context.Context, pgDSN string, onNew func(ctx context.Context, a classes.Attempt) error) error {
	pool, err := pgxpool.New(ctx, pgDSN)
	if err != nil {
		fmt.Printf("pgxpool.New: %w", err)
		return err
	}
	defer pool.Close()

	// Берём отдельное подключение для LISTEN (чтобы уведомления не “пропадали” среди запросов пула)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		fmt.Printf("acquire conn: %w", err)
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, `LISTEN attempt_insert`)
	if err != nil {
		fmt.Printf("LISTEN failed: %w", err)
		return err
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
			fmt.Printf("wait notification: %w", err)
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

		// передаём дальше в твою бизнес-логику (Executor/обработчик)
		if err := onNew(ctx, att); err != nil {
			log.Printf("processing attempt %s failed: %v", id, err)
		}
	}
}

func loadAttemptByID(ctx context.Context, pool *pgxpool.Pool, id string) (classes.Attempt, error) {
	var a classes.Attempt

	// под твой текущий struct (id, created_at, … programming_language_id/name, verdicts)
	row := pool.QueryRow(ctx, `
		SELECT
			id,
			created_at,
			git_student_url,
			git_site_url,
			variable_with_url,
			task_id,
			task_name,
			programming_language_id,
			programming_language_name,
			testing_verdict,
			postmoderation
		FROM attempt
		WHERE id = $1
	`, id)

	err := row.Scan(
		&a.Id,
		&a.CreatedAt,
		&a.GitStudentURL,
		&a.GitSiteURL,
		&a.VariableWithURL,
		&a.TaskId,
		&a.TaskName,
		&a.ProgrammingLanguageId,
		&a.ProgrammingLanguageName,
		&a.TestingVerdict,
		&a.Postmoderation,
	)
	return a, err
}
