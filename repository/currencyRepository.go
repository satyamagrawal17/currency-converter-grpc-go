package repository

import (
	"context"
	"currency_converter1/database"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

// LocalDynamoDBConfig provides configuration for local DynamoDB development
type CurrencyRepository struct {
	db        *dynamodb.Client
	tableName string
}

func NewCurrencyRepository(dy *database.DynamoDB) (ICurrencyRepository, error) {
	return &CurrencyRepository{
		db:        dy.DB,
		tableName: dy.TableName,
	}, nil
}

func (dbRepo *CurrencyRepository) UpdateItems(newRates map[string]float64) error {
	for currency, rate := range newRates {
		input := &dynamodb.UpdateItemInput{
			TableName: aws.String(dbRepo.tableName),
			Key: map[string]types.AttributeValue{
				"Currency": &types.AttributeValueMemberS{Value: currency},
			},
			UpdateExpression: aws.String("SET Rate = :r"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":r": &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", rate)},
			},
		}

		_, err := dbRepo.db.UpdateItem(context.TODO(), input)
		if err != nil {
			return fmt.Errorf("failed to update item (%s -> %f): %w", currency, rate, err)
		}
	}
	return nil
}

func (dbRepo *CurrencyRepository) GetItem(currency string) (float64, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(dbRepo.tableName),
		Key: map[string]types.AttributeValue{
			"Currency": &types.AttributeValueMemberS{Value: currency},
		},
	}

	result, err := dbRepo.db.GetItem(context.TODO(), input)
	if err != nil {
		return 0, fmt.Errorf("failed to get item for currency %s: %w", currency, err)
	}

	if result.Item == nil {
		return 0, fmt.Errorf("currency %s not found", currency)
	}

	rateStr := result.Item["Rate"].(*types.AttributeValueMemberN).Value
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rate for currency %s: %w", currency, err)
	}

	return rate, nil
}
