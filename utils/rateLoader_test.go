package utils

import (
	"os"
	"testing"
)

func TestLoadRates(t *testing.T) {
	// Create a temporary file with JSON content
	file, err := os.CreateTemp("", "rates*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	jsonContent := `{"USD": 1.0, "EUR": 0.85}`
	if _, err := file.Write([]byte(jsonContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	file.Close()

	// Test the loadRates function
	rates, err := loadRates(file.Name())
	if err != nil {
		t.Fatalf("loadRates returned an error: %v", err)
	}

	// Verify the loaded rates
	expectedRates := map[string]float64{"USD": 1.0, "EUR": 0.85}
	for currency, rate := range expectedRates {
		if rates[currency] != rate {
			t.Errorf("Expected rate for %s: %f, got: %f", currency, rate, rates[currency])
		}
	}
}
