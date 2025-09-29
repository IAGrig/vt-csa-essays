package mocks

import (
    "context"

    "github.com/IAGrig/vt-csa-essays/backend/review-service/internal/kafka"
    "github.com/stretchr/testify/mock"
)

type MockProducer struct {
    mock.Mock
}

func (m *MockProducer) SendNotificationEvent(ctx context.Context, event kafka.NotificationEvent) error {
    args := m.Called(ctx, event)
    return args.Error(0)
}

func (m *MockProducer) Close() error {
    args := m.Called()
    return args.Error(0)
}
