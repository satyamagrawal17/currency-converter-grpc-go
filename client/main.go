package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "currency_converter1/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "tcp", addr)
	}))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCurrencyConversionClient(conn)

	req := &pb.CurrencyConversionRequest{
		Money: &pb.Money{
			Currency: "USD",
			Amount:   100,
		},
		FromCurrency: "EUR",
	}

	r, err := c.ConvertCurrency(ctx, req)
	if err != nil {
		log.Fatalf("could not convert currency: %v", err)
	}
	log.Printf("Converted Amount: %f %s", r.Money.Amount, r.Money.Currency)
}
