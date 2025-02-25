package utils

import (
	"encoding/json"
	"os"
)

func loadRates(filename string) (map[string]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rates map[string]float64
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&rates)
	if err != nil {
		return nil, err
	}
	return rates, nil
}
