package service

import (
	"context"
	"fmt"
	"wb-test-task/internal/cache"
	"wb-test-task/internal/db"
	"wb-test-task/internal/models"
)

type OrderService struct {
	repo  *db.Repository
	cache *cache.LRUCache
}

func NewOrderService(repo *db.Repository, cache *cache.LRUCache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

// Функция ProcessOrder сохраняет заказ в БД и кэше
func (s *OrderService) ProcessOrder(ctx context.Context, order models.Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("Пустой UID заказа")
	}

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("Не удалось сохранить заказ: %w", err)
	}

	// s.cache.Set(order.OrderUID, &order) // Добавляем в кэш
	return nil
}

// Получение заказа из кэша или БД
func (s *OrderService) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	// Проверяем, есть ли заказ в кэше
	if cached, exists := s.cache.Get(orderUID); exists {
		fmt.Println("Заказ найден в кэше")

		return cached.(*models.Order), nil // Возвращаем заказ из кэша
	}

	// Если заказа нет в кэше, получаем из БД и добавляем в кэш
	order, err := s.repo.GetOrder(ctx, orderUID)
	if err != nil {
		return nil, fmt.Errorf("Ошибка получения заказа из БД: %w", err)
	}
	s.cache.Set(order.OrderUID, &order)

	fmt.Println("Заказ найден в БД")

	if err != nil {
		return nil, fmt.Errorf("Ошибка получения заказа из БД: %w", err)
	}

	s.cache.Set(orderUID, order)
	return order, nil
}
