package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wb-test-task/internal/handlers"
	"wb-test-task/internal/models"
	"wb-test-task/internal/service"
)

// We reuse mockRepo/mockCache from the previous file.

func TestGetOrderByUID_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := &mockRepo{getOrderRes: &models.Order{OrderUID: "uid-1"}}
	cache := newMockCache() 
	svc := service.NewOrderService(repo, cache)
	h := handlers.NewOrderHandler(svc)

	r.GET("/order/:orderId", h.GetOrderByUID)

	req := httptest.NewRequest(http.MethodGet, "/order/uid-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var got models.Order
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "uid-1", got.OrderUID)
}

func TestGetOrderByUID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := &mockRepo{getOrderErr: context.Canceled}
	cache := newMockCache()
	svc := service.NewOrderService(repo, cache)
	h := handlers.NewOrderHandler(svc)

	r.GET("/order/:orderId", h.GetOrderByUID)

	req := httptest.NewRequest(http.MethodGet, "/order/nope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}
