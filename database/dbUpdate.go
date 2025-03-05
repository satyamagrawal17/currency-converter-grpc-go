package database

import (
	"currency_converter1/config"
	"database/sql"
	"log"
	"sync"
)

// Database represents a database connection or data storage.
type CurrencyDB struct {
	rates map[string]float64
	mu    sync.Mutex
	db    *sql.DB
}

func NewDatabase() (*CurrencyDB, error) {
	cfg, err := config.LoadConfig()
	connStr := buildConnStr(cfg)             // Replace with your connection string
	db, err := sql.Open("postgres", connStr) // or mysql, mongo etc.
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &CurrencyDB{
		db:    db,
		rates: make(map[string]float64),
	}
	err = database.loadInitialData()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (d *CurrencyDB) loadInitialData() error {
	rows, err := d.db.Query("SELECT currency, rate FROM conversions") // Replace with your query
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var currency string
		var rate float64
		if err := rows.Scan(&currency, &rate); err != nil {
			return err
		}
		d.rates[currency] = rate
	}
	return nil
}

// UpdateDB updates rates in the database.
func (d *CurrencyDB) UpdateDB(newRates map[string]float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update the rates in the database
	for currency, rate := range newRates {
		if _, exists := d.rates[currency]; exists {
			_, err := d.db.Exec("UPDATE conversions SET rate = $1 WHERE currency = $2", rate, currency) // Replace with your update query
			if err != nil {
				log.Printf("Error updating database: %v\n", err)
				return err
			}
			d.rates[currency] = rate
		}
	}
	log.Println("Database rates updated.")
	return nil
}
