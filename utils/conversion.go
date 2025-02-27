package utils

import (
	"database/sql"
	"fmt"
)

type ConversionUtils struct{}

func NewConversionUtils() *ConversionUtils {
	return &ConversionUtils{}
}

func (u *ConversionUtils) GetConversionRate(db *sql.DB, fromCurrency, toCurrency string) (float64, error) {
	var rate float64
	query := "SELECT rate FROM conversions WHERE from_currency = $1 AND to_currency = $2"
	err := db.QueryRow(query, fromCurrency, toCurrency).Scan(&rate)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("conversion rate not found for %s to %s", fromCurrency, toCurrency)
		}
		return 0, fmt.Errorf("error querying conversion rate: %v", err)
	}
	return rate, nil
}
