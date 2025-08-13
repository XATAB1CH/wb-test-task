package handlers

import (
	"encoding/json"
	"net/http"
	"wb-test-task/internal/service"
)

func NewRouter(svc *service.OrderService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/order/", getOrderHandler(svc))
	return mux
}

// Хендлер для получения заказа по UID
func getOrderHandler(svc *service.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderUID := r.URL.Path[len("/order/"):]
		if orderUID == "" {
			http.Error(w, "Требуется корректны UID", http.StatusBadRequest)
			return
		}

		order, err := svc.GetOrder(r.Context(), orderUID)
		if err != nil {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		respondWithJSON(w, http.StatusOK, order)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
