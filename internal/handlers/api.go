package handlers

import (
	"encoding/json"
	"net/http"
	"wb-test-task/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Хендлер получения заказа по UID
func (h *OrderHandler) GetOrderByUID(c *gin.Context) {
	ctx := c.Request.Context() // достаём контекст из запроса

	uid := c.Param("orderId")

	order, err := h.svc.GetOrder(ctx, uid)
	if err != nil { // Если возникла ошибка, возвращаем HTTP статус 500 и сообщение об ошибке
		respondWithJSON(c.Writer, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Если все хорошо, возвращаем HTTP статус 200 и заказ
	respondWithJSON(c.Writer, http.StatusOK, order)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
