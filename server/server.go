package main

import (
	"currency_converter1/database"
	"currency_converter1/pb"
	"currency_converter1/service"
	"currency_converter1/utils"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	// Connect to the database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize utility instance
	conversionUtils := utils.NewConversionUtils()

	// Create CurrencyConverterService with initialized DB and Utils
	converterService := &service.CurrencyConverterService{
		DB:    db,
		Utils: conversionUtils,
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCurrencyConversionServer(s, converterService)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
