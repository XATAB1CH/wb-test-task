package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"wb-test-task/config"
	"wb-test-task/internal/bootstrap"
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

	// Шардированный кэш (numShards, capacity, ttl)
	lru := cache.NewShardedLRU[*models.Order](32, cfg.CacheCapacity, cfg.CacheTTL)

	// Сервис поверх интерфейсов
	svc := service.NewOrderService(repo, lru)

	// Восстановление кэша из БД — отдельно от интерфейсов
	if err := bootstrap.RestoreCacheFromDB(ctx, repo, lru); err != nil {
		log.Fatalf("cache restore: %v", err)
	}

	// HTTP
	router := gin.Default()
	router.Static("/assets", "./internal/assets")
	router.LoadHTMLGlob("internal/templates/*")
	routes.InitRoutes(router, svc)

	srv := &http.Server{
		Addr:              ":" + webPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Kafka consumer (актуальная сигнатура)
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	groupID := cfg.KafkaGroupID
	topic := cfg.KafkaTopic
	consumer := kafka.NewConsumer(brokers, groupID, topic, repo, lru)

	// Старт consumer
	go func() { consumer.Run(ctx) }()

	// Старт HTTP
	go func() {
		log.Printf("HTTP listening on :%s", webPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown: signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}

	// consumer закрывается сам по отмене ctx внутри Run
	log.Println("graceful shutdown complete")
}
