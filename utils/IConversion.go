package utils

import "database/sql"

type IConversion interface {
	GetConversionRate(db *sql.DB, fromCurrency, toCurrency string) (float64, error)
}
