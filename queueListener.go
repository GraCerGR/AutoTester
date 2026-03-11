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

			threads_number,
			threads_reuse,

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
		&a.Threads.Number,
		&a.Threads.Reuse,
		&a.ShutdownCondition,
	)
	return a, err
}
