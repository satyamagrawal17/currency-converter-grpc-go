package server

import (
	"context"
	"currency_converter1/consumer"
	"currency_converter1/pb"
	"currency_converter1/repository"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func StartSignalHandler(ctx context.Context, cancel context.CancelFunc) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		log.Println("Shutting down...")
		cancel()
	}()
}

func StartConsumer(ctx context.Context, dbRepo repository.ICurrencyRepository) {
	go consumer.ConsumeMessages(ctx, dbRepo)
}

func StartGrpcGateway(ctx context.Context) {
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
}
