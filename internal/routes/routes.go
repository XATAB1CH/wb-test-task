package routes

import (
	"net/http"
	"wb-test-task/internal/handlers"
	"wb-test-task/internal/service"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, svc *service.OrderService) *gin.Engine {

	h := handlers.NewOrderHandler(svc)

	api := r.Group("/") // инициализируем роутер для API
	{
		api.GET("/", func(c *gin.Context) { // хендлер для главной страницы
			c.HTML(http.StatusOK, "index.html", nil)
		})
	}

	order := r.Group("/order") // инициализируем роутер для заказов
	{
		order.GET("/:orderId", h.GetOrderByUID) // хендлер для поиска заказа по UID
	}

	return r
}
