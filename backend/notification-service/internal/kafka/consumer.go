package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/repository"
	"github.com/IAGrig/vt-csa-essays/backend/shared/monitoring"

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
}

func NewConsumer(brokers []string, topic string, groupID string, repo repository.NotificationRepository) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
})

	return &Consumer{
		reader:     reader,
		repository: repo,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer...")
			c.reader.Close()
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
		log.Printf("Error fetching message: %v", err)
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "fetch_error").Inc()
		return
	}

	var event NotificationEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "unmarshall_error").Inc()
		return
	}

	notificationReq := models.NotificationRequest{
		UserID:  event.UserID,
		Content: event.Content,
	}

	_, err = c.repository.Create(notificationReq)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "creating_error").Inc()
		return
	}

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "commit_error").Inc()
		log.Printf("Error committing message: %v", err)
		return
	}

	duration := time.Since(start)
	monitoring.DbQueryDuration.WithLabelValues("create", "notifications").Observe(float64(duration.Milliseconds()))
	monitoring.KafkaMessagesProcessed.WithLabelValues("notifications", "success").Inc()
	monitoring.NotificationsCreated.WithLabelValues(event.Author).Inc()
}
