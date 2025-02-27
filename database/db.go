package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}

func buildConnStr() string {
	user := getEnv("DB_USER")
	password := getEnv("DB_PASSWORD")
	dbname := getEnv("DB_NAME")
	host := getEnv("DB_HOST")
	port := getEnv("DB_PORT")

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", user, password, dbname, host, port)
}

func createTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS conversions (
			from_currency VARCHAR(3) NOT NULL,
			to_currency VARCHAR(3) NOT NULL,
			rate DECIMAL NOT NULL,
			PRIMARY KEY (from_currency, to_currency)
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
			fromCurrency string
			toCurrency   string
			rate         float64
		}{
			{"USD", "EUR", 0.92},
			{"EUR", "USD", 1.09},
			{"USD", "INR", 83.5},
			{"INR", "USD", 0.012},
			{"EUR", "INR", 91.0},
			{"INR", "EUR", 0.011},
		}

		for _, sample := range samples {
			_, err := db.Exec("INSERT INTO conversions (from_currency, to_currency, rate) VALUES ($1, $2, $3) ON CONFLICT (from_currency, to_currency) DO NOTHING", sample.fromCurrency, sample.toCurrency, sample.rate)
			if err != nil {
				log.Printf("Error inserting sample (%s -> %s): %v", sample.fromCurrency, sample.toCurrency, err)
			}
		}
	}
}

func ConnectDB() (*sql.DB, error) {
	loadEnv()

	connStr := buildConnStr()
	log.Printf(connStr)

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
