package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"
	"wb-test-task/internal/db"
)

type LRUCache struct { // LRU значит Least Recently Used - наименее используемый в данный момент
	capacity int                      // количество элементов в кэше
	ttl      time.Duration            // время жизни элемента в кэше
	items    map[string]*list.Element // map для хранения элементов
	list     *list.List               // список для хранения элементов в порядке использования
	mu       sync.Mutex               // мьютекс для синхронизации доступа к кэшу
}

type cacheItem struct {
	key       string      // ключ - orderUID
	value     interface{} // значение - *models.Order
	expiresAt time.Time   // время истечения срока действия
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*list.Element), // инициализируем map
		list:     list.New(),                     // инициализируем список
	}
}

// Добавляем элемент в кэш или обновляем существующий
func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists { // если элемент уже существует, обновляем его
		c.list.MoveToFront(elem) // перемещаем элемент в начало списка
		elem.Value.(*cacheItem).value = value
		elem.Value.(*cacheItem).expiresAt = time.Now().Add(c.ttl)
		return
	}

	if c.list.Len() >= c.capacity { // если кэш заполнен, удаляем самый старый элемент
		oldest := c.list.Back()
		delete(c.items, oldest.Value.(*cacheItem).key)
		c.list.Remove(oldest)
	}

	item := &cacheItem{ // создаем новый элемент
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}

	elem := c.list.PushFront(item) // добавляем элемент в начало списка
	c.items[key] = elem            // добавляем элемент в map
}

// Восстановка кэша из БД
func (c *LRUCache) Restore(ctx context.Context, repo *db.Repository) error {
	// Получаем все элементы из БД
	orders, err := repo.GetAllOrders(ctx)
	if err != nil {
		return fmt.Errorf("Не удалось восстановить кэш из БД: %w", err)
	}

	// Записываем их в кэш
	for _, order := range orders {
		c.Set(order.OrderUID, &order)
	}

	return nil
}

// Получаем элемент из кэша по ключу
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(elem.Value.(*cacheItem).expiresAt) { // если элемент устарел, удаляем его
		c.list.Remove(elem)
		delete(c.items, key)
		return nil, false
	}

	c.list.MoveToFront(elem)
	return elem.Value.(*cacheItem).value, true
}
