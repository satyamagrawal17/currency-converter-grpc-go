package service

import (
	"currency_converter1/dto"
	"currency_converter1/repository"
	"fmt"
)

type CurrencyConverterService struct {
	repository repository.ICurrencyRepository
}

func NewCurrencyConverterService(currencyRepo repository.ICurrencyRepository) *CurrencyConverterService {
	return &CurrencyConverterService{
		repository: currencyRepo,
	}
}

func (s *CurrencyConverterService) ConvertCurrency(req dto.ConvertCurrencyRequest) (float64, error) {

	fromRate, err := s.getConversionRate(req.FromCurrency)
	if err != nil {
		return 0, fmt.Errorf("could not get conversion rate for %s: %v", req.FromCurrency, err)
	}

	toRate, err := s.getConversionRate(req.Money.Currency)
	if err != nil {
		return 0, fmt.Errorf("could not get conversion rate for %s: %v", req.Money.Currency, err)
	}

	convertedAmount := (req.Money.Amount * toRate) / fromRate
	return convertedAmount, nil
}

func (s *CurrencyConverterService) getConversionRate(currency string) (float64, error) {
	rate, err := s.repository.GetItem(currency)
	if err != nil {
		return 0, fmt.Errorf("failed to get conversion rate for currency %s: %w", currency, err)
	}
	return rate, nil
}
