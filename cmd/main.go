package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wb-test-task/config"
	"wb-test-task/internal/cache"
	"wb-test-task/internal/db"
	"wb-test-task/internal/kafka"
	"wb-test-task/internal/models"
	"wb-test-task/internal/service"
	"wb-test-task/internal/handlers"
)

func main() {
	cfg := config.LoadConfig() // загружаем данные окружения

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // установка таймаута на 10 секунд
	defer cancel()

	pgPool, err := db.NewPostgresPool(ctx, cfg) // подключение к БД
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	repo := db.NewRepository(pgPool) // инициализация репозитория
	cache := cache.NewLRUCache(1000, 10*time.Minute) // инициализация кэша
	svc := service.NewOrderService(repo, cache) // инициализация сервиса

	kafkaConsumer := kafka.NewConsumer(cfg, func(order models.Order) error { // второй переменной передаётся функция, которая будет вызываться при получении сообщения. это функция обработчик
		return svc.ProcessOrder(context.Background(), order)
	})
	defer kafkaConsumer.Close()

	go func() {
		if err := kafkaConsumer.Consume(context.Background()); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()
	
	// поднимаем сервер
	server := &http.Server{
		Addr:    ":8081",
		Handler: handlers.NewRouter(svc),
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
