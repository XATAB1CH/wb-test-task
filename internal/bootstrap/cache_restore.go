package bootstrap

import (
	"context"
	"fmt"

	"wb-test-task/internal/db"
	"wb-test-task/internal/models"
	"wb-test-task/internal/ports"
)

// RestoreCacheFromDB загружает заказы из БД и кладёт их в кэш.
// Специально использует конкретную реализацию repo (*db.Repository),
// чтобы не расширять интерфейс ports.OrderRepository.
func RestoreCacheFromDB(ctx context.Context, repo *db.Repository, c ports.Cache[string, *models.Order]) error {
	orders, err := repo.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("restore cache: get all orders: %w", err)
	}

	for i := range orders {
		c.Set(orders[i].OrderUID, &orders[i])
	}
	return nil
}
