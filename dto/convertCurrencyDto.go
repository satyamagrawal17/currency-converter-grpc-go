package dto

import (
	"currency_converter1/models"
)

type ConvertCurrencyRequest struct {
	Money        models.Money
	FromCurrency string
}
