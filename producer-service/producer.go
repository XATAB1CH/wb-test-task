package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

const (
	topic              = "orders"
	kafkaBrokerAddress = "localhost:29093"
)

func main() {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaBrokerAddress},
		Topic:   topic,
	})
	defer writer.Close()

	// читаем JSON файлы из папки
	ordersDir, err := os.ReadDir("./orders")
	if err != nil {
		log.Fatal(err)
	}

	sendMsgsToKafka(writer, ordersDir)

	fmt.Println("Продюсер отправил файлы!")
}

func sendMsgsToKafka(writer *kafka.Writer, dir []os.DirEntry) {

	for i, orderJSON := range dir {
		if orderJSON.IsDir() {
			continue
		}

		// Читаем JSON файл
		jsonData, err := os.ReadFile(fmt.Sprintf("./orders/%s", orderJSON.Name()))
		if err != nil {
			log.Fatalf("Ошибка при чтении файла №%v: %v", i, err)
		}

		// Проверяем, что данные являются валидным JSON
		if !json.Valid(jsonData) {
			log.Fatalf("Файл №%v не содержит валидный JSON", i)
		}

		msg := kafka.Message{
			Value: jsonData,
			Time:  time.Now(),
		}

		if err := writer.WriteMessages(context.Background(), msg); err != nil {
			log.Printf("Сообщение №%v не отправлено в Kafka: %s\n", i, err)
		} else {
			fmt.Printf("Сообщение №%v отправлено в Kafka!\n", i)
		}

		time.Sleep(200 * time.Millisecond)
	}
}
