package service

import (
	"context"
	"currency_converter1/pb"
	"currency_converter1/utils"
	"database/sql"
	"fmt"
)

type CurrencyConverterService struct {
	pb.UnimplementedCurrencyConversionServer
	DB    *sql.DB
	Utils utils.IConversion
}

func (s *CurrencyConverterService) ConvertCurrency(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.CurrencyConversionResponse, error) {
	if req.FromCurrency == "" {
		return nil, fmt.Errorf("from currency cannot be empty")
	}
	if req.Money.Currency == "" {
		return nil, fmt.Errorf("to currency cannot be empty")
	}
	toRate, err := s.Utils.GetConversionRate(s.DB, req.Money.Currency)
	if err != nil {
		return nil, fmt.Errorf("could not get conversion rate: %v", err)
	}

	fromRate, err := s.Utils.GetConversionRate(s.DB, req.FromCurrency)
	if err != nil {
		return nil, fmt.Errorf("could not get conversion rate: %v", err)
	}

	convertedAmount := (req.Money.Amount * toRate) / fromRate
	money := &pb.Money{
		Currency: req.Money.Currency,
		Amount:   convertedAmount,
	}
	return &pb.CurrencyConversionResponse{Money: money}, nil
}
