package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"refina-transaction/config/log"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/utils/data"

	"github.com/rabbitmq/amqp091-go"
)

// TransactionCreator is an interface to decouple the consumer from the service package.
type TransactionCreator interface {
	CreateTransaction(ctx context.Context, transaction dto.TransactionsRequest) (dto.TransactionsResponse, error)
}

type InvestmentEventConsumer struct {
	rabbitMQ           RabbitMQClient
	transactionCreator TransactionCreator
}

func NewInvestmentEventConsumer(rmq RabbitMQClient, txnCreator TransactionCreator) *InvestmentEventConsumer {
	return &InvestmentEventConsumer{
		rabbitMQ:           rmq,
		transactionCreator: txnCreator,
	}
}

// investmentBuyEvent represents the payload published by investment-service on investment.buy
type investmentBuyEvent struct {
	ID               string  `json:"id"`
	UserID           string  `json:"userId"`
	Code             string  `json:"code"`
	Quantity         float64 `json:"quantity"`
	InitialValuation float64 `json:"initialValuation"`
	Amount           float64 `json:"amount"`
	Date             string  `json:"date"`
	Description      string  `json:"description"`
	WalletID         string  `json:"walletId"`
}

// investmentSellEvent represents a single sold record from investment-service on investment.sell
type investmentSellEvent struct {
	ID           string  `json:"id"`
	UserID       string  `json:"userId"`
	InvestmentID string  `json:"investmentId"`
	Quantity     float64 `json:"quantity"`
	SellPrice    float64 `json:"sellPrice"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Description  string  `json:"description"`
	Deficit      float64 `json:"deficit"`
	WalletID     string  `json:"walletId"`
}

func (c *InvestmentEventConsumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info(data.LogInvestmentConsumerStopped, map[string]any{"service": data.InvestmentConsumerService})
			return
		default:
			if err := c.consume(ctx); err != nil {
				log.Error(data.LogInvestmentConsumerFailed, map[string]any{
					"service": data.InvestmentConsumerService,
					"error":   err.Error(),
				})
				// Wait before reconnecting
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
			}
		}
	}
}

func (c *InvestmentEventConsumer) consume(ctx context.Context) error {
	channel, err := c.rabbitMQ.GetChannel()
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}
	defer channel.Close()

	// Declare the queue for investment events
	queue, err := channel.QueueDeclare(
		"transaction.investment.events", // queue name
		true,                            // durable
		false,                           // auto-delete
		false,                           // exclusive
		false,                           // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	// Bind queue to investment.buy routing key
	if err := channel.QueueBind(queue.Name, data.EVENT_INVESTMENT_BUY, data.OUTBOX_PUBLISH_EXCHANGE, false, nil); err != nil {
		return fmt.Errorf("bind queue investment.buy: %w", err)
	}

	// Bind queue to investment.sell routing key
	if err := channel.QueueBind(queue.Name, data.EVENT_INVESTMENT_SELL, data.OUTBOX_PUBLISH_EXCHANGE, false, nil); err != nil {
		return fmt.Errorf("bind queue investment.sell: %w", err)
	}

	msgs, err := channel.Consume(
		queue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Info(data.LogInvestmentConsumerStarted, map[string]any{"service": data.InvestmentConsumerService})

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			if err := c.handleMessage(ctx, msg); err != nil {
				log.Error(data.LogInvestmentEventHandleFailed, map[string]any{
					"service":     data.InvestmentConsumerService,
					"routing_key": msg.RoutingKey,
					"error":       err.Error(),
				})
				// Nack and requeue
				_ = msg.Nack(false, true)
			} else {
				_ = msg.Ack(false)
			}
		}
	}
}

func (c *InvestmentEventConsumer) handleMessage(ctx context.Context, msg amqp091.Delivery) error {
	switch msg.RoutingKey {
	case data.EVENT_INVESTMENT_BUY:
		return c.handleInvestmentBuy(ctx, msg.Body)
	case data.EVENT_INVESTMENT_SELL:
		return c.handleInvestmentSell(ctx, msg.Body)
	default:
		log.Warn(data.LogInvestmentEventUnknown, map[string]any{
			"service":     data.InvestmentConsumerService,
			"routing_key": msg.RoutingKey,
		})
		return nil
	}
}

func (c *InvestmentEventConsumer) handleInvestmentBuy(ctx context.Context, body []byte) error {
	var event investmentBuyEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("unmarshal investment buy event: %w", err)
	}

	if event.WalletID == "" {
		log.Warn(data.LogInvestmentEventSkipped, map[string]any{
			"service": data.InvestmentConsumerService,
			"reason":  "wallet_id is empty",
			"event":   "investment.buy",
		})
		return nil
	}

	investmentDate, err := time.Parse(time.RFC3339, event.Date)
	if err != nil {
		// Try alternative date format
		investmentDate, err = time.Parse("2006-01-02T15:04:05.000Z", event.Date)
		if err != nil {
			return fmt.Errorf("parse date: %w", err)
		}
	}

	assetCode := event.Code
	description := fmt.Sprintf("Pembelian investasi %s sebanyak %.4f", assetCode, event.Quantity)

	txnReq := dto.TransactionsRequest{
		WalletID:           event.WalletID,
		CategoryID:         data.CATEGORY_ID_INVESTMENT_BUY,
		Amount:             event.Amount,
		Date:               investmentDate,
		Description:        description,
		Attachments:        []dto.UpdateAttachmentsRequest{},
		IsWalletNotCreated: false,
	}

	_, err = c.transactionCreator.CreateTransaction(ctx, txnReq)
	if err != nil {
		return fmt.Errorf("create buy transaction: %w", err)
	}

	log.Info(data.LogInvestmentBuyTransactionCreated, map[string]any{
		"service":       data.InvestmentConsumerService,
		"investment_id": event.ID,
		"wallet_id":     event.WalletID,
		"amount":        event.Amount,
	})

	return nil
}

func (c *InvestmentEventConsumer) handleInvestmentSell(ctx context.Context, body []byte) error {
	var events []investmentSellEvent
	if err := json.Unmarshal(body, &events); err != nil {
		return fmt.Errorf("unmarshal investment sell events: %w", err)
	}

	for _, event := range events {
		if event.WalletID == "" {
			log.Warn(data.LogInvestmentEventSkipped, map[string]any{
				"service": data.InvestmentConsumerService,
				"reason":  "wallet_id is empty",
				"event":   "investment.sell",
			})
			continue
		}

		investmentDate, err := time.Parse(time.RFC3339, event.Date)
		if err != nil {
			investmentDate, err = time.Parse("2006-01-02T15:04:05.000Z", event.Date)
			if err != nil {
				log.Error(data.LogInvestmentEventHandleFailed, map[string]any{
					"service": data.InvestmentConsumerService,
					"sold_id": event.ID,
					"error":   fmt.Sprintf("parse date: %s", err.Error()),
				})
				continue
			}
		}

		description := fmt.Sprintf("Penjualan investasi sebanyak %.4f dengan harga jual %.0f/unit", event.Quantity, event.SellPrice)

		txnReq := dto.TransactionsRequest{
			WalletID:           event.WalletID,
			CategoryID:         data.CATEGORY_ID_INVESTMENT_SELL,
			Amount:             event.Amount,
			Date:               investmentDate,
			Description:        description,
			Attachments:        []dto.UpdateAttachmentsRequest{},
			IsWalletNotCreated: false,
		}

		_, err = c.transactionCreator.CreateTransaction(ctx, txnReq)
		if err != nil {
			log.Error(data.LogInvestmentEventHandleFailed, map[string]any{
				"service": data.InvestmentConsumerService,
				"sold_id": event.ID,
				"error":   err.Error(),
			})
			continue
		}

		log.Info(data.LogInvestmentSellTransactionCreated, map[string]any{
			"service":   data.InvestmentConsumerService,
			"sold_id":   event.ID,
			"wallet_id": event.WalletID,
			"amount":    event.Amount,
		})
	}

	return nil
}
