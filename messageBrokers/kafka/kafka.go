package messageBrokers

import (
	"MainApp/classes"
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

func PublishAttempt(ctx context.Context, w *kafka.Writer, a classes.Attempt) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(a.ProgrammingLanguageName),
		Value: data,
		Time:  time.Now(),
	})
}