package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/IAGrig/vt-csa-essays/backend/notification-service/internal/models"
	"github.com/IAGrig/vt-csa-essays/backend/shared/pg_util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationPgRepository struct {
	db *pgxpool.Pool
}

func NewNotificationPgRepository() (NotificationRepository, error) {
	pool, err := pgutil.GetPgxPool()
	if err != nil {
		return nil, err
	}

	return &NotificationPgRepository{db: pool}, nil
}

func (repository *NotificationPgRepository) Create(request models.NotificationRequest) (models.Notification, error) {
	var n models.Notification
	err := repository.db.QueryRow(context.Background(),
		`INSERT INTO notifications (user_id, content)
		VALUES ($1, $2)
		RETURNING notification_id, is_read, created_at;`,
		request.UserID,
		request.Content,
	).Scan(&n.NotificationID, &n.IsRead, &n.CreatedAt)

	if err != nil {
		return models.Notification{}, fmt.Errorf("failed to create notification: %w", err)
	}

	n.UserID = request.UserID
	n.Content = request.Content

	return n, nil
}

func (repository *NotificationPgRepository) GetByUserID(userID int64) ([]models.Notification, error) {
	rows, err := repository.db.Query(context.Background(),
		`SELECT notification_id, user_id, content, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC;`,
		userID,
	)
	if err != nil {
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
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

func (repository *NotificationPgRepository) MarkAsRead(notificationID int64) error {
	result, err := repository.db.Exec(context.Background(),
		`UPDATE notifications
		SET is_read = true
		WHERE notification_id = $1;`,
		notificationID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return NotificationNotFoundErr
	}

	return nil
}

func (repository *NotificationPgRepository) MarkAllAsRead(userID int64) error {
	_, err := repository.db.Exec(context.Background(),
		`UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false;`,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

func (repository *NotificationPgRepository) GetByID(notificationID int64) (models.Notification, error) {
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
			return models.Notification{}, NotificationNotFoundErr
		}
		return models.Notification{}, fmt.Errorf("failed to get notification: %w", err)
	}

	return n, nil
}

func (repository *NotificationPgRepository) DB() *pgxpool.Pool {
	return repository.db
}
