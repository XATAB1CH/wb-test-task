package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"wb-test-task/config"
	"wb-test-task/internal/cache"
	"wb-test-task/internal/db"
	"wb-test-task/internal/kafka"
	"wb-test-task/internal/models"
	"wb-test-task/internal/routes"
	"wb-test-task/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Подключение к PostgreSQL
	pgPool, err := db.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	repo := db.NewRepository(pgPool)

	// Инициализация кэша
	myCache := cache.NewLRUCache(1000, 10*time.Minute)

	// Восстанавливаем кэш из БД
	if err := myCache.Restore(ctx, repo); err != nil {
		log.Fatalf("Failed to restore cache: %v", err)
	}

	fmt.Println(myCache)

	// Сервис заказов
	svc := service.NewOrderService(repo, myCache)

	// Kafka consumer
	kafkaConsumer := kafka.NewConsumer(cfg, func(order models.Order) error {
		return svc.ProcessOrder(context.Background(), order)
	})
	defer kafkaConsumer.Close()

	go func() {
		if err := kafkaConsumer.Consume(context.Background()); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	// Создаём Gin router
	router := gin.Default()
	router.LoadHTMLGlob("./internal/templates/*") // загрузка шаблонов

	routes.InitRoutes(router, svc) // подключаем маршруты

	// Канал для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	go func() {
		if err := router.Run(":8081"); err != nil {
			log.Fatalf("Gin server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")
}
