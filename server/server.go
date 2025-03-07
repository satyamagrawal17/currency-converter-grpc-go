package main

import (
	"context"
	"currency_converter1/consumer"
	"currency_converter1/database"
	"currency_converter1/pb"
	"currency_converter1/repository"
	"currency_converter1/service"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	dbInstance, err := database.InitDynamoDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	dbRepo, err := repository.NewCurrencyRepository(dbInstance)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	converterService, err := service.NewCurrencyConverterService(dbRepo)
	if err != nil {
		log.Fatalf("failed to connect to service: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterCurrencyConversionServer(s, converterService)

	ctx, cancel := context.WithCancel(context.Background())
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		log.Println("Shutting down...")
		cancel()
	}()

	go consumer.ConsumeMessages(ctx, dbRepo)

	go func() {
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err := pb.RegisterCurrencyConversionHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
		if err != nil {
			log.Fatalf("failed to start gRPC-Gateway server: %v", err)
		}
		log.Println("gRPC-Gateway server started. Listening on :8081")
		if err := http.ListenAndServe(":8081", mux); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
