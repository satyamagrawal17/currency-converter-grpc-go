package main

import (
	"context"
	"currency_converter1/database"
	"currency_converter1/pb"
	"currency_converter1/repository"
	"currency_converter1/server"
	"currency_converter1/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	dbInstance, err := database.InitDynamoDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	dbRepo := repository.NewCurrencyRepository(dbInstance)
	converterService := service.NewCurrencyConverterService(dbRepo)
	grpcServer := server.NewGrpcServer(converterService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCurrencyConversionServer(s, grpcServer)

	ctx, cancel := context.WithCancel(context.Background())

	server.StartSignalHandler(ctx, cancel)
	server.StartConsumer(ctx, dbRepo)
	server.StartGrpcGateway(ctx)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
