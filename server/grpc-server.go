package server

import (
	"context"
	"currency_converter1/dto"
	"currency_converter1/models"
	"currency_converter1/pb"
	"currency_converter1/service"
	"fmt"
)

type GrpcServer struct {
	pb.UnimplementedCurrencyConversionServer
	conversionService *service.CurrencyConverterService
}

func NewGrpcServer(service *service.CurrencyConverterService) *GrpcServer {
	return &GrpcServer{
		conversionService: service,
	}
}

func (s *GrpcServer) ConvertCurrency(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.CurrencyConversionResponse, error) {
	if req.FromCurrency == "" {
		return nil, fmt.Errorf("from currency cannot be empty")
	}
	if req.Money.Currency == "" {
		return nil, fmt.Errorf("to currency cannot be empty")
	}
	if req.Money.Amount <= 0 {
		return nil, fmt.Errorf("amount should be positove only")
	}
	savedMoney := models.Money{
		Amount:   req.Money.Amount,
		Currency: req.Money.Currency,
	}
	conversionRequestDto := dto.ConvertCurrencyRequest{
		Money:        savedMoney,
		FromCurrency: req.FromCurrency,
	}
	convertedAmount, err := s.conversionService.ConvertCurrency(conversionRequestDto)
	if err != nil {
		return nil, fmt.Errorf("could not convert amount")
	}
	convertedMoney := &pb.Money{
		Currency: req.Money.Currency,
		Amount:   convertedAmount,
	}
	return &pb.CurrencyConversionResponse{Money: convertedMoney}, nil
}
