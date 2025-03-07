package service

import (
	"context"
	"currency_converter1/pb"
	"currency_converter1/repository"
	"fmt"
	"github.com/golang/mock/gomock"
	"math"
	"testing"
)

func setupMockController(t *testing.T) (*gomock.Controller, *repository.MockDynamoDBRepository, *CurrencyConverterService) {
	ctrl := gomock.NewController(t)
	mockRepo := repository.NewMockDynamoDBRepository(ctrl)
	s := &CurrencyConverterService{DB: mockRepo}
	return ctrl, mockRepo, s
}

func TestConvertCurrency_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		fromCurrency   string
		toCurrency     string
		amount         float64
		expectedAmount float64
		expectError    bool
		mockSetup      func(mockRepo *repository.MockDynamoDBRepository)
	}{
		{
			name:           "Same currency conversion",
			fromCurrency:   "USD",
			toCurrency:     "USD",
			amount:         100,
			expectedAmount: 100,
			expectError:    false,
			mockSetup: func(mockRepo *repository.MockDynamoDBRepository) {
				mockRepo.EXPECT().GetItem("USD").Return(1.0, nil).Times(2)
			},
		},
		{
			name:           "Very large amount",
			fromCurrency:   "USD",
			toCurrency:     "EUR",
			amount:         1e9,
			expectedAmount: 9.2e8,
			expectError:    false,
			mockSetup: func(mockRepo *repository.MockDynamoDBRepository) {
				mockRepo.EXPECT().GetItem("USD").Return(1.0, nil)
				mockRepo.EXPECT().GetItem("EUR").Return(0.92, nil)
			},
		},
		{
			name:           "Very small amount",
			fromCurrency:   "USD",
			toCurrency:     "EUR",
			amount:         1e-9,
			expectedAmount: 9.2e-10,
			expectError:    false,
			mockSetup: func(mockRepo *repository.MockDynamoDBRepository) {
				mockRepo.EXPECT().GetItem("USD").Return(1.0, nil)
				mockRepo.EXPECT().GetItem("EUR").Return(0.92, nil)
			},
		},
		{
			name:           "Non-existent currency",
			fromCurrency:   "USD",
			toCurrency:     "ABC",
			amount:         100,
			expectedAmount: 0,
			expectError:    true,
			mockSetup: func(mockRepo *repository.MockDynamoDBRepository) {
				mockRepo.EXPECT().GetItem("USD").Return(1.0, nil)
				mockRepo.EXPECT().GetItem("ABC").Return(0.0, fmt.Errorf("currency ABC not found"))
			},
		},
		{
			name:           "Empty from currency",
			fromCurrency:   "",
			toCurrency:     "USD",
			amount:         100,
			expectedAmount: 0,
			expectError:    true,
			mockSetup:      func(mockRepo *repository.MockDynamoDBRepository) {},
		},
		{
			name:           "Empty to currency",
			fromCurrency:   "USD",
			toCurrency:     "",
			amount:         100,
			expectedAmount: 0,
			expectError:    true,
			mockSetup:      func(mockRepo *repository.MockDynamoDBRepository) {},
		},
		{
			name:           "Zero rate",
			fromCurrency:   "USD",
			toCurrency:     "JPY",
			amount:         0,
			expectedAmount: 0,
			expectError:    false,
			mockSetup: func(mockRepo *repository.MockDynamoDBRepository) {
				mockRepo.EXPECT().GetItem("USD").Return(1.0, nil)
				mockRepo.EXPECT().GetItem("JPY").Return(0.0, nil)
			},
		},
	}

	const tolerance = 1e-12

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl, mockRepo, s := setupMockController(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

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
