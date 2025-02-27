package service

import (
	"context"
	"currency_converter1/pb"
	"database/sql"
	"fmt"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type MockUtils struct {
	GetConversionRateFunc func(db *sql.DB, fromCurrency, toCurrency string) (float64, error)
}

func (m *MockUtils) GetConversionRate(db *sql.DB, fromCurrency, toCurrency string) (float64, error) {
	return m.GetConversionRateFunc(db, fromCurrency, toCurrency)
}

func TestConvertCurrency_EdgeCases(t *testing.T) {
	// Mock database connection
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	// Mock GetConversionRate function
	mockUtils := &MockUtils{
		GetConversionRateFunc: func(db *sql.DB, fromCurrency, toCurrency string) (float64, error) {
			rates := map[string]float64{
				"USD": 1.0,
				"EUR": 0.85,
				"JPY": 110.0,
			}
			rate, ok := rates[toCurrency]
			if !ok {
				return 0, fmt.Errorf("conversion rate not found for %s to %s", fromCurrency, toCurrency)
			}
			return rate, nil
		},
	}

	s := &CurrencyConverterService{DB: db, Utils: mockUtils}

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
				Money: &pb.Money{
					Currency: tt.toCurrency,
					Amount:   tt.amount,
				},
				FromCurrency: tt.fromCurrency,
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
