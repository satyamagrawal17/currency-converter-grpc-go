package main

import (
	"context"
	"currency_converter1/consumer"
	"currency_converter1/database"
	"currency_converter1/pb"
	"currency_converter1/service"
	"currency_converter1/utils"
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
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	conversionUtils := utils.NewConversionUtils()

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

	currencyDB, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("failed to create CurrencyDB: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		log.Println("Shutting down...")
		cancel()
	}()

	go consumer.ConsumeMessages(ctx, currencyDB)

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
