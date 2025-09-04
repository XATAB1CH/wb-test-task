package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	webPort := os.Getenv("WEB_PORT")
	if webPort == "" {
		webPort = "8081"
	}

	// Контекст жизни приложения: отменится по SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("postgres pool: %v", err)
	}
	defer pool.Close()

	repo := db.NewRepository(pool)

	lru := cache.NewLRUCache(cfg.CacheCapacity, cfg.CacheTTL)

	svc := service.NewOrderService(repo, lru)

	router := gin.Default()
	router.Static("/assets", "./internal/assets")
	router.LoadHTMLGlob("internal/templates/*")
	routes.InitRoutes(router, svc)

	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: router,
	}

	// Kafka consumer: пробрасываем общий ctx и обработчик
	consumer := kafka.NewConsumer(cfg, func(o models.Order) error {
		// используем общий контекст приложения — он отменится при shutdown
		return svc.SaveOrder(ctx, o)
	})

	// Восстанавливаем кэш из БД
	if err := lru.Restore(ctx, repo); err != nil {
		log.Fatalf("Failed to restore cache: %v", err)
	}

	// Старт consumer в отдельной горутине
	go func() {
		if err := consumer.Consume(ctx); err != nil && ctx.Err() == nil {
			log.Printf("kafka consumer stopped with error: %v", err)
		}
	}()

	// Старт HTTP-сервера
	go func() {
		log.Printf("HTTP listening on :%s", webPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown: signal received")

	// Плавное завершение: даём время активным запросам и горутинам
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()

	// Останоливаем HTTP server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	// Закрываем consumer (разбудит блокирующие операции, Consume выйдет по ctx)
	if err := consumer.Close(); err != nil {
		log.Printf("kafka close: %v", err)
	}

	log.Println("graceful shutdown complete")
}
