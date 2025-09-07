package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wb-test-task/internal/models"
	"wb-test-task/internal/service"
)

// Моки для тестов
type mockRepo struct {
	lastUID     string
	gotCtx      context.Context
	getOrderRes *models.Order
	getOrderErr error
	callsGet    int
}

func (m *mockRepo) SaveOrder(ctx context.Context, o models.Order) error {
	return nil
}

func (m *mockRepo) GetOrder(ctx context.Context, uid string) (*models.Order, error) {
	m.callsGet++
	m.gotCtx = ctx
	m.lastUID = uid
	return m.getOrderRes, m.getOrderErr
}

type mockCache struct {
	store    map[string]*models.Order
	callsGet int
	callsSet int
	callsDel int
}

func newMockCache() *mockCache {
	return &mockCache{store: map[string]*models.Order{}}
}

func (m *mockCache) Get(key string) (*models.Order, bool) {
	m.callsGet++
	v, ok := m.store[key]
	return v, ok
}

func (m *mockCache) Set(key string, value *models.Order) {
	m.callsSet++
	m.store[key] = value
}

func (m *mockCache) Delete(key string) { m.callsDel++; delete(m.store, key) }

// Тесты
func TestOrderService_GetOrderByUID_CacheHit(t *testing.T) {
	repo := &mockRepo{}
	cache := newMockCache()
	want := &models.Order{OrderUID: "uid-1"}
	cache.store["uid-1"] = want

	svc := service.NewOrderService(repo, cache)
	got, err := svc.GetOrderByUID(context.Background(), "uid-1")
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, 0, repo.callsGet, "repo.GetOrder не вызывать при попадании в кэш")
}

func TestOrderService_GetOrderByUID_CacheMiss_Success(t *testing.T) {
	repo := &mockRepo{getOrderRes: &models.Order{OrderUID: "uid-2"}}
	cache := newMockCache()

	svc := service.NewOrderService(repo, cache)
	got, err := svc.GetOrderByUID(context.Background(), "uid-2")
	require.NoError(t, err)
	assert.Equal(t, "uid-2", got.OrderUID)
	assert.Equal(t, 1, repo.callsGet)

	_, ok := cache.store["uid-2"]
	assert.True(t, ok)
	assert.Equal(t, 1, cache.callsSet)
}
