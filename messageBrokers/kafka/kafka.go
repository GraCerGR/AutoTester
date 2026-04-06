package messagebrokers

import (
	"MainApp/classes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func PublishAttemptResult(ctx context.Context, writer *kafka.Writer, result classes.AllTestsInChecker, attemptID uuid.UUID) error {
	b, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(attemptID.String()),
		Value: b,
	})
}

func CreateTopicKafka(URL, topic string, numPartitions, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", URL)
	if err != nil {
		return fmt.Errorf("Ошибка подключения к Kafka: %w", err)
	}

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}

	if err := conn.CreateTopics(topicConfig); err != nil {
		return fmt.Errorf("Ошибка создания topic: %w", err)
	}
	defer conn.Close()
	return nil
}
