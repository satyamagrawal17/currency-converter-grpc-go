package service

import (
	"context"
	"currency_converter1/pb"
)

type ICurrencyConverterService interface {
	ConvertCurrency(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.CurrencyConversionResponse, error)
}
