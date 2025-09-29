package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type NotificationEvent struct {
	Type     string `json:"type"`
	UserID   int64  `json:"user_id"`
	Content  string `json:"content"`
	EssayID  int64  `json:"essay_id"`
	ReviewID int64  `json:"review_id"`
	Author   string `json:"author"`
}

type Producer interface {
	SendNotificationEvent(ctx context.Context, event NotificationEvent) error
	Close() error
}

type KafkaProducer struct {
	writer *kafka.Writer
}


func NewProducer(brokers []string, topic string) Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
	}
}

func (p *KafkaProducer) SendNotificationEvent(ctx context.Context, event NotificationEvent) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: eventBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("Sent notification event for user %d", event.UserID)
	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
