package consumer

import (
	"context"
	"currency_converter1/config"
	"currency_converter1/database"
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func ConsumeMessages(ctx context.Context, currencyDB *database.CurrencyDB) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("failed to fetch environment variables: %v\n", err)
		return // Return early if config loading fails
	}
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBroker,
		"group.id":          cfg.KafkaGroupId,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{cfg.KafkaTopic}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %s\n", err)
	}

	log.Println("Kafka consumer started. Listening on topic:", cfg.KafkaTopic)

	run := true

	go func() {
		<-ctx.Done()
		log.Println("Shutting down consumer...")
		run = false
	}()

	for run {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrTimedOut {
				continue
			}
			log.Printf("Error reading message: %v\n", err)
			continue
		}

		log.Printf("Received message: key = %s, value = %s, partition = %d, offset = %d\n", string(msg.Key), string(msg.Value), msg.TopicPartition.Partition, msg.TopicPartition.Offset)

		var kafkaMsg config.KafkaMessage
		err = json.Unmarshal(msg.Value, &kafkaMsg)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v\n", err)
			continue
		}

		err = currencyDB.UpdateDB(kafkaMsg.Rates)
		if err != nil {
			log.Printf("Error updating database: %v\n", err)
			continue
		}

		_, err = consumer.CommitMessage(msg)
		if err != nil {
			log.Printf("Error committing message: %v\n", err)
		}
	}

	log.Println("Consumer shutdown complete.")
}
