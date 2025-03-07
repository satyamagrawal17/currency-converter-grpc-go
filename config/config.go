package configure

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type KafkaMessage struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int64              `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float64 `json:"rates"`
}

type Config struct {
	DBUser          string
	DBPassword      string
	DBName          string
	DBHost          string
	DBPort          string
	KafkaBroker     string
	KafkaTopic      string
	KafkaGroupId    string
	DynamoEndpoint  string
	DynamoRegion    string
	DynamoAccessKey string
	DynamoSecretKey string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{
		DBUser:          os.Getenv("DB_USER"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		DBHost:          os.Getenv("DB_HOST"),
		DBPort:          os.Getenv("DB_PORT"),
		KafkaBroker:     os.Getenv("KAFKA_BROKER"),
		KafkaTopic:      os.Getenv("KAFKA_TOPIC"),
		KafkaGroupId:    os.Getenv("GROUP_ID"),
		DynamoEndpoint:  os.Getenv("DYNAMO_ENDPOINT"),
		DynamoRegion:    os.Getenv("DYNAMO_REGION"),
		DynamoAccessKey: os.Getenv("DYNAMO_ACCESS_KEY"),
		DynamoSecretKey: os.Getenv("DYNAMO_SECRET_KEY"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" || cfg.DBHost == "" || cfg.DBPort == "" || cfg.KafkaBroker == "" || cfg.KafkaTopic == "" || cfg.KafkaGroupId == "" || cfg.DynamoEndpoint == "" || cfg.DynamoRegion == "" || cfg.DynamoAccessKey == "" || cfg.DynamoSecretKey == "" {
		return nil, fmt.Errorf("one or more environment variables are not set")
	}

	return cfg, nil
}
