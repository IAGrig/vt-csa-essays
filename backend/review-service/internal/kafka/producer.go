package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"go.uber.org/zap"

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
	logger *logging.Logger
}

func NewProducer(brokers []string, topic string, logger *logging.Logger) Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
		logger: logger,
	}
}

func (p *KafkaProducer) SendNotificationEvent(ctx context.Context, event NotificationEvent) error {
	logger := p.logger.With(
		zap.String("operation", "send_notification_event"),
		zap.String("type", event.Type),
		zap.Int64("user_id", event.UserID),
		zap.Int64("essay_id", event.EssayID),
		zap.Int64("review_id", event.ReviewID),
		zap.String("author", event.Author),
	)

	logger.Debug("Sending notification event to Kafka")

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal notification event", zap.Error(err))
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: eventBytes,
	})
	if err != nil {
		logger.Error("Failed to write message to Kafka", zap.Error(err))
		return fmt.Errorf("failed to write message: %w", err)
	}

	logger.Info("Notification event sent successfully")
	return nil
}

func (p *KafkaProducer) Close() error {
	p.logger.Debug("Closing Kafka producer")
	return p.writer.Close()
}
