package repository

type ICurrencyRepository interface {
	GetItem(currency string) (float64, error)
	UpdateItems(newRates map[string]float64) error
}
