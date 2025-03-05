package utils

import "database/sql"

type IConversion interface {
	GetConversionRate(db *sql.DB, currency string) (float64, error)
}
