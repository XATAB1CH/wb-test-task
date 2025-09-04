package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	HTTPPort            string
	HTTPShutdownTimeout time.Duration

	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string

	CacheCapacity int
	CacheTTL      time.Duration
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// cache default
	viper.SetDefault("CACHE_CAPACITY", 1000)
	viper.SetDefault("CACHE_TTL", "5m")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Ошибка чтения .env файла: %v", err)
		return nil, err
	}

	return &Config{
		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),

		KafkaBrokers: viper.GetString("KAFKA_BROKERS"),
		KafkaTopic:   viper.GetString("KAFKA_TOPIC"),
		KafkaGroupID: viper.GetString("KAFKA_GROUP_ID"),

		HTTPPort:            viper.GetString("HTTP_PORT"),
		HTTPShutdownTimeout: time.Duration(viper.GetInt("HTTP_SHUTDOWNTIMEOUT_SEC")) * time.Second,

		CacheCapacity: viper.GetInt("CACHE_CAPACITY"),
		CacheTTL:      time.Duration(viper.GetInt("CACHE_TTL_MIN")) * time.Minute,
	}, nil
}
