package models

type Rate struct {
	Currency string  `dynamodbav:"Currency"`
	Rate     float64 `dynamodbav:"Rate"`
}
