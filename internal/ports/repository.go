package ports

import (
	"context"
	"wb-test-task/internal/models"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, o models.Order) error
	GetOrder(ctx context.Context, uid string) (*models.Order, error)
	// PreloadCache(ctx context.Context) ([]models.Order, error) // больше не нужен
}
