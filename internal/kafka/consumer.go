package kafka

import (
	"context"
	"encoding/json"
	"fmt"
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
			GroupID:  cfg.KafkaGroupID,
			MinBytes: 1e3, // не уверен, нужно ли добавлять в конфиг
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

			// Обработка (сохранение в БД)
			if err := c.handler(order); err != nil {
				log.Printf("Ошибка обработки заказа %s: %v", order.OrderUID, err)
				// ❌ оффсет не коммитим → сообщение придет снова
				continue
			}

			// ✅ только если всё прошло успешно → коммитим оффсет
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Ошибка при коммите оффсета: %v", err)
			}

			fmt.Printf("JSON файл %s обработан!\n", order.OrderUID)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
