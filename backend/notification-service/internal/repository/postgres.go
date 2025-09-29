package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/logging"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationPgRepository struct {
	db     *pgxpool.Pool
	logger *logging.Logger
}

func NewNotificationPgRepository(logger *logging.Logger) (NotificationRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &NotificationPgRepository{db: pool, logger: logger}, nil
}

func (repository *NotificationPgRepository) Create(request models.NotificationRequest) (models.Notification, error) {
	logger := repository.logger.With(
		zap.String("operation", "create_notification"),
		zap.Int64("user_id", request.UserID),
	)

	logger.Debug("Creating notification")

	var n models.Notification
	err := repository.db.QueryRow(context.Background(),
		`INSERT INTO notifications (user_id, content)
		VALUES ($1, $2)
		RETURNING notification_id, is_read, created_at;`,
		request.UserID,
		request.Content,
	).Scan(&n.NotificationID, &n.IsRead, &n.CreatedAt)

	if err != nil {
		logger.Error("Failed to create notification in database", zap.Error(err))
		return models.Notification{}, fmt.Errorf("failed to create notification: %w", err)
	}

	n.UserID = request.UserID
	n.Content = request.Content

	logger.Info("Notification created successfully",
		zap.Int64("notification_id", n.NotificationID))
	return n, nil
}

func (repository *NotificationPgRepository) GetByUserID(userID int64) ([]models.Notification, error) {
	logger := repository.logger.With(
		zap.String("operation", "get_notifications_by_user_id"),
		zap.Int64("user_id", userID),
	)

	logger.Debug("Getting notifications by user ID")

	rows, err := repository.db.Query(context.Background(),
		`SELECT notification_id, user_id, content, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC;`,
		userID,
	)
	if err != nil {
		logger.Error("Failed to get notifications from database", zap.Error(err))
		return nil, fmt.Errorf("failed to load notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		err = rows.Scan(
			&n.NotificationID,
			&n.UserID,
			&n.Content,
			&n.IsRead,
			&n.CreatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan notification row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	logger.Debug("Retrieved notifications", zap.Int("count", len(notifications)))
	return notifications, nil
}

func (repository *NotificationPgRepository) MarkAsRead(notificationID int64) error {
	logger := repository.logger.With(
		zap.String("operation", "mark_notification_as_read"),
		zap.Int64("notification_id", notificationID),
	)

	logger.Debug("Marking notification as read")

	result, err := repository.db.Exec(context.Background(),
		`UPDATE notifications
		SET is_read = true
		WHERE notification_id = $1;`,
		notificationID,
	)

	if err != nil {
		logger.Error("Failed to mark notification as read in database", zap.Error(err))
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		logger.Warn("Notification not found for marking as read")
		return NotificationNotFoundErr
	}

	logger.Debug("Notification marked as read successfully")
	return nil
}

func (repository *NotificationPgRepository) MarkAllAsRead(userID int64) error {
	logger := repository.logger.With(
		zap.String("operation", "mark_all_notifications_as_read"),
		zap.Int64("user_id", userID),
	)

	logger.Debug("Marking all notifications as read for user")

	result, err := repository.db.Exec(context.Background(),
		`UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false;`,
		userID,
	)

	if err != nil {
		logger.Error("Failed to mark all notifications as read in database", zap.Error(err))
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	rowsAffected := result.RowsAffected()
	logger.Debug("Marked notifications as read", zap.Int64("rows_affected", rowsAffected))
	return nil
}

func (repository *NotificationPgRepository) GetByID(notificationID int64) (models.Notification, error) {
	logger := repository.logger.With(
		zap.String("operation", "get_notification_by_id"),
		zap.Int64("notification_id", notificationID),
	)

	logger.Debug("Getting notification by ID")

	var n models.Notification
	err := repository.db.QueryRow(context.Background(),
		`SELECT notification_id, user_id, content, is_read, created_at
		FROM notifications
		WHERE notification_id = $1;`,
		notificationID,
	).Scan(
		&n.NotificationID,
		&n.UserID,
		&n.Content,
		&n.IsRead,
		&n.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("Notification not found")
			return models.Notification{}, NotificationNotFoundErr
		}
		logger.Error("Failed to get notification from database", zap.Error(err))
		return models.Notification{}, fmt.Errorf("failed to get notification: %w", err)
	}

	logger.Debug("Notification retrieved successfully")
	return n, nil
}

func (repository *NotificationPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
