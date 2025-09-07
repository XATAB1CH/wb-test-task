package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"wb-test-task/config"
	"wb-test-task/internal/bootstrap"
	wbcache "wb-test-task/internal/cache"
	"wb-test-task/internal/db"
	"wb-test-task/internal/kafka"
	"wb-test-task/internal/models"
	"wb-test-task/internal/routes"
	"wb-test-task/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("postgres pool: %v", err)
	}
	defer pool.Close()

	repo := db.NewRepository(pool)

	cache := wbcache.NewShardedLRU[*models.Order](16, cfg.CacheCapacity, cfg.CacheTTL)
	if err := bootstrap.RestoreCacheFromDB(ctx, repo, cache); err != nil {
		log.Printf("bootstrap cache: %v", err)
	}

	svc := service.NewOrderService(repo, cache)

	r := gin.Default()
	r.Static("/assets", "./internal/assets")
	r.LoadHTMLGlob("internal/templates/*")
	r = routes.InitRoutes(r, svc)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	consumer := kafka.NewConsumer(brokers, cfg.KafkaGroupID, cfg.KafkaTopic, repo, cache)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
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

	wg.Wait()
	log.Println("graceful shutdown complete")
}
