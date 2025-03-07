package service

import (
	"context"
	"currency_converter1/pb"
	"currency_converter1/repository"
	"fmt"
)

type CurrencyConverterService struct {
	pb.UnimplementedCurrencyConversionServer
	DB repository.ICurrencyRepository
}

func NewCurrencyConverterService(currencyRepo repository.ICurrencyRepository) (*CurrencyConverterService, error) {
	return &CurrencyConverterService{
		DB: currencyRepo,
	}, nil
}

func (s *CurrencyConverterService) ConvertCurrency(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.CurrencyConversionResponse, error) {
	if req.FromCurrency == "" {
		return nil, fmt.Errorf("from currency cannot be empty")
	}
	if req.Money.Currency == "" {
		return nil, fmt.Errorf("to currency cannot be empty")
	}

	fromRate, err := s.getConversionRate(req.FromCurrency)
	if err != nil {
		return nil, fmt.Errorf("could not get conversion rate for %s: %v", req.FromCurrency, err)
	}

	toRate, err := s.getConversionRate(req.Money.Currency)
	if err != nil {
		return nil, fmt.Errorf("could not get conversion rate for %s: %v", req.Money.Currency, err)
	}

	convertedAmount := (req.Money.Amount * toRate) / fromRate
	money := &pb.Money{
		Currency: req.Money.Currency,
		Amount:   convertedAmount,
	}
	return &pb.CurrencyConversionResponse{Money: money}, nil
}

func (s *CurrencyConverterService) getConversionRate(currency string) (float64, error) {
	rate, err := s.DB.GetItem(currency)
	if err != nil {
		return 0, fmt.Errorf("failed to get conversion rate for currency %s: %w", currency, err)
	}
	return rate, nil
}
