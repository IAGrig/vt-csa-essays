package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"
	"go.uber.org/zap"

	"github.com/segmentio/kafka-go"
)

type NotificationEvent struct {
	UserID    int64  `json:"user_id"`
	Content   string `json:"content"`
	EssayID   int64  `json:"essay_id,omitempty"`
	ReviewID  int64  `json:"review_id,omitempty"`
	Author    string `json:"author,omitempty"`
}

type Consumer struct {
	reader     *kafka.Reader
	repository repository.NotificationRepository
	logger     *logging.Logger
}

func NewConsumer(brokers []string, topic string, groupID string, repo repository.NotificationRepository, logger *logging.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader:     reader,
		repository: repo,
		logger:     logger,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("Starting Kafka consumer",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group_id", c.reader.Config().GroupID))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping Kafka consumer...")
			if err := c.reader.Close(); err != nil {
				c.logger.Error("Error closing Kafka reader", zap.Error(err))
			}
			return
		default:
			c.consumeMessage(ctx)
		}
	}
}

func (c *Consumer) consumeMessage(ctx context.Context) {
	start := time.Now()

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		c.logger.Error("Error fetching Kafka message", zap.Error(err))
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "fetch_error").Inc()
		return
	}

	c.logger.Debug("Received Kafka message",
		zap.String("topic", msg.Topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset))

	var event NotificationEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("Error unmarshaling Kafka message",
			zap.Error(err),
			zap.ByteString("message", msg.Value))
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "unmarshall_error").Inc()
		return
	}

	logger := c.logger.With(
		zap.Int64("user_id", event.UserID),
		zap.String("author", event.Author),
	)

	logger.Debug("Processing notification event")

	notificationReq := models.NotificationRequest{
		UserID:  event.UserID,
		Content: event.Content,
	}

	notification, err := c.repository.Create(notificationReq)
	if err != nil {
		logger.Error("Error creating notification from Kafka event",
			zap.Error(err))
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "creating_error").Inc()
		return
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		logger.Error("Error committing Kafka message",
			zap.Error(err),
			zap.Int64("notification_id", notification.NotificationID))
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "commit_error").Inc()
		return
	}

	duration := time.Since(start)
	monitoring.DbQueryDuration.WithLabelValues("create", "notifications").Observe(float64(duration.Milliseconds()))
	monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "success").Inc()
	monitoring.NotificationsCreated.WithLabelValues(event.Author).Inc()

	logger.Info("Notification created from Kafka event",
		zap.Int64("notification_id", notification.NotificationID),
		zap.Duration("processing_time", duration))
}
