package service

import (
	"context"
	"currency_converter1/pb"
	"currency_converter1/utils"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedCurrencyConversionServer
	rates map[string]float64
}

func (s *server) ConvertCurrency(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.CurrencyConversionResponse, error) {
	// Implement your conversion logic here
	fromCurrentRate, exists := s.rates[req.FromCurrency]
	if !exists {
		return nil, fmt.Errorf("conversion rate not found for %s", req.FromCurrency)
	}
	toCurrentRate, exists := s.rates[req.ToCurrency]
	if !exists {
		return nil, fmt.Errorf("conversion rate not found for %s", req.ToCurrency)
	}
	rate := toCurrentRate / fromCurrentRate
	convertedAmount := req.Amount * rate
	money := &pb.Money{
		Currency: req.ToCurrency,
		Amount:   convertedAmount,
	}
	return &pb.CurrencyConversionResponse{Money: money}, nil
}

func main() {
	rates, err := utils.LoadRates("config/currencies.json")
	if err != nil {
		log.Fatalf("failed to load rates: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCurrencyConversionServer(s, &server{rates: rates})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
