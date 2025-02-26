package main

import (
	"context"
	"currency_converter1/pb"
	"math"
	"testing"
)

func TestConvertCurrency_EdgeCases(t *testing.T) {
	s := &server{
		rates: map[string]float64{
			"USD": 1.0,
			"EUR": 0.85,
			"JPY": 110.0,
		},
	}

	tests := []struct {
		name           string
		fromCurrency   string
		toCurrency     string
		amount         float64
		expectedAmount float64
		expectError    bool
	}{
		{"Same currency conversion", "USD", "USD", 100, 100, false},
		{"Very large amount", "USD", "EUR", 1e9, 8.5e8, false},
		{"Very small amount", "USD", "EUR", 1e-9, 8.5e-10, false},
		{"Non-existent currency", "USD", "ABC", 100, 0, true},
		{"Empty from currency", "", "USD", 100, 0, true},
		{"Empty to currency", "USD", "", 100, 0, true},
		{"Zero rate", "USD", "JPY", 0, 0, false},
	}

	const tolerance = 1e-12

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.CurrencyConversionRequest{
				FromCurrency: tt.fromCurrency,
				ToCurrency:   tt.toCurrency,
				Amount:       tt.amount,
			}

			resp, err := s.ConvertCurrency(context.Background(), req)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got: %v", tt.expectError, err)
			}

			if !tt.expectError {
				if resp.Money.Currency != tt.toCurrency {
					t.Errorf("expected currency %s, got %s", tt.toCurrency, resp.Money.Currency)
				}
				if math.Abs(resp.Money.Amount-tt.expectedAmount) > tolerance {
					t.Errorf("expected amount %f, got %f", tt.expectedAmount, resp.Money.Amount)
				}
			}
		})
	}
}
