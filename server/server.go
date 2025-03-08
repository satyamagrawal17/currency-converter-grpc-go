package server

import (
	"context"
	"currency_converter1/database"
	"currency_converter1/pb"
	"currency_converter1/repository"
	"currency_converter1/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func Start() {
	dbInstance, err := database.InitDynamoDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	dbRepo := repository.NewCurrencyRepository(dbInstance)
	converterService := service.NewCurrencyConverterService(dbRepo)
	grpcServer := NewGrpcServer(converterService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCurrencyConversionServer(s, grpcServer)

	ctx, cancel := context.WithCancel(context.Background())

	StartSignalHandler(ctx, cancel)
	StartConsumer(ctx, dbRepo)
	StartGrpcGateway(ctx)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
