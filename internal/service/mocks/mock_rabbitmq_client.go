package mocks

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type MockRabbitMQClient struct {
	mock.Mock
}

func (m *MockRabbitMQClient) GetChannel() (*amqp091.Channel, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amqp091.Channel), args.Error(1)
}

func (m *MockRabbitMQClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitMQClient) Publish(ctx context.Context, routingKey string, body []byte) error {
	args := m.Called(ctx, routingKey, body)
	return args.Error(0)
}
