package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"wb-test-task/config"
	"wb-test-task/internal/models"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	handler func(models.Order) error
}

func NewConsumer(cfg *config.Config, handler func(models.Order) error) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{cfg.KafkaBrokers},
			Topic:    cfg.KafkaTopic,
			MinBytes: 1e3,
			MaxBytes: 1e6,
		}),
		handler: handler,
	}
}

func (c *Consumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx) // читаем сообщение из Kafka
			if err != nil {
				log.Printf("Не получилось прочитать сообщение из Kafka: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			var order models.Order // декодируем сообщение в структуру Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Не получилось декодировать JSON: %v", err)
				continue
			}

			if err := c.handler(order); err != nil {
				log.Printf("Ошибка декодирования заказа %s: %v", order.OrderUID, err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
