package database

import (
	"currency_converter1/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

//type Config struct {
//	APIURL       string
//	KafkaBroker  string
//	KafkaTopic   string
//	APIKey       string
//	CronSchedule string
//}
//
//func loadEnv() {
//	err := godotenv.Load()
//	if err != nil {
//		log.Fatalf("Error loading .env file: %v", err)
//	}
//}
//
//func getEnv(key string) string {
//	value := os.Getenv(key)
//	if value == "" {
//		log.Fatalf("Environment variable %s not set", key)
//	}
//	return value
//}

func buildConnStr(cfg *config.Config) string {
	user := cfg.DBUser
	password := cfg.DBPassword
	dbname := cfg.DBName
	host := cfg.DBHost
	port := cfg.DBPort

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", user, password, dbname, host, port)
}

func createTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS conversions (
			currency VARCHAR(3) NOT NULL,
			rate DECIMAL NOT NULL,
			PRIMARY KEY (currency)
		);
	`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func insertSamples(db *sql.DB) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM conversions").Scan(&count)
	if err != nil {
		log.Fatalf("Error checking table count: %v", err)
	}

	if count == 0 {
		samples := []struct {
			currency string
			rate     float64
		}{
			{"USD", 1},
			{"EUR", 1.09},
			{"INR", 83.5},
		}

		for _, sample := range samples {
			_, err := db.Exec("INSERT INTO conversions (currency, rate) VALUES ($1, $2) ON CONFLICT (currency) DO NOTHING", sample.currency, sample.rate)
			if err != nil {
				log.Printf("Error inserting sample (%s -> %s): %v", sample.currency, err)
			}
		}
	}
}

func ConnectDB() (*sql.DB, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch environment variables: %v", err)
	}

	connStr := buildConnStr(cfg)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Database connection successful!")

	createTable(db)
	insertSamples(db)

	return db, nil
}
