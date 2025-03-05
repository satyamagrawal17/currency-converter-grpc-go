package utils

import (
	"database/sql"
	"fmt"
)

type ConversionUtils struct{}

func NewConversionUtils() *ConversionUtils {
	return &ConversionUtils{}
}

func (u *ConversionUtils) GetConversionRate(db *sql.DB, currency string) (float64, error) {
	var rate float64
	query := "SELECT rate FROM conversions WHERE currency = $1"
	err := db.QueryRow(query, currency).Scan(&rate)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversion rate not found for %s", currency)
		}
		return 0, fmt.Errorf("error querying conversion rate: %v", err)
	}
	return rate, nil
}
