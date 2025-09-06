package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"wb-test-task/internal/models"
	"wb-test-task/internal/ports"
	"wb-test-task/internal/validation"
)

type Consumer struct {
	reader *kafka.Reader
	repo   ports.OrderRepository
	cache  ports.Cache[string, *models.Order]
}

func NewConsumer(brokers []string, groupID, topic string, repo ports.OrderRepository, cache ports.Cache[string, *models.Order]) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	return &Consumer{
		reader: r,
		repo:   repo,
		cache:  cache,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	log.Printf("[kafka] consumer started (group=%q)", c.reader.Config().GroupID)
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("[kafka] reader close: %v", err)
		}
		log.Printf("[kafka] consumer stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf("[kafka] fetch canceled: %v", err)
				return
			}
			log.Printf("[kafka] fetch: %v", err)
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if err := c.processMessage(ctx, msg); err != nil {
			log.Printf("[kafka] process: %v", err)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var order models.Order

	if err := json.Unmarshal(msg.Value, &order); err != nil {
		log.Printf("[kafka] bad json (partition=%d, offset=%d): %v", msg.Partition, msg.Offset, err)
		if err2 := c.reader.CommitMessages(ctx, msg); err2 != nil {
			log.Printf("[kafka] commit after bad json: %v", err2)
		}
		return nil
	}

	if err := validation.ValidateStruct(order); err != nil {
		log.Printf("[kafka] validation failed (uid=%s, offset=%d): %v", order.OrderUID, msg.Offset, err)
		if err2 := c.reader.CommitMessages(ctx, msg); err2 != nil {
			log.Printf("[kafka] commit after validation error: %v", err2)
		}
		return nil
	}

	if err := c.repo.SaveOrder(ctx, order); err != nil {
		log.Printf("[kafka] db save failed (uid=%s, offset=%d): %v", order.OrderUID, msg.Offset, err)
		time.Sleep(300 * time.Millisecond)
		return nil
	}

	c.cache.Set(order.OrderUID, &order)

	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		log.Printf("[kafka] commit offset failed (uid=%s, offset=%d): %v", order.OrderUID, msg.Offset, err)
	}
	return nil
}
