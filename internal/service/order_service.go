package service

import (
	"context"
	"fmt"
	"wb-test-task/internal/models"
	"wb-test-task/internal/ports"
)

type OrderService struct {
	repo  ports.OrderRepository
	cache ports.Cache[string, *models.Order]
}

func NewOrderService(repo ports.OrderRepository, cache ports.Cache[string, *models.Order]) *OrderService {
	return &OrderService{repo: repo, cache: cache}
}

func (s *OrderService) GetOrderByUID(ctx context.Context, orderUID string) (*models.Order, error) {
	if order, ok := s.cache.Get(orderUID); ok && order != nil {
		return order, nil
	}

	order, err := s.repo.GetOrder(ctx, orderUID)
	if err != nil {
		return nil, fmt.Errorf("get order from db: %w", err)
	}
	s.cache.Set(orderUID, order)
	return order, nil
}
