package repository

import (
	"context"
	configure "currency_converter1/config"
	"currency_converter1/models"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
)

// LocalDynamoDBConfig provides configuration for local DynamoDB development
type CurrencyRepository struct {
	db        *dynamodb.Client
	tableName string
}

func NewClient() (*CurrencyRepository, error) {
	cfg, err := configure.LoadConfig()

	// Create custom resolver for local DynamoDB
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           cfg.DynamoEndpoint,
			SigningRegion: cfg.DynamoRegion,
		}, nil
	})

	// Load AWS configuration with local settings
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.DynamoRegion),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.DynamoAccessKey,
			cfg.DynamoSecretKey,
			"local-session",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load local DynamoDB config: %w", err)
	}

	// Create DynamoDB client
	newDbInstance := &CurrencyRepository{
		db:        dynamodb.NewFromConfig(awsConfig),
		tableName: "conversions",
	}
	if err := newDbInstance.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	if err := newDbInstance.initializeTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}
	return newDbInstance, nil
}

func (dbRepo *CurrencyRepository) createTable() error {
	// Check if the table already exists
	exists, err := dbRepo.doesTableExists()
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if exists {
		// Table already exists, no need to create
		return nil
	}

	// Define the table schema
	input := &dynamodb.CreateTableInput{
		TableName: &dbRepo.tableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("Currency"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("Currency"),
				KeyType:       types.KeyTypeHash,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	}

	// Create the table
	_, err = dbRepo.db.CreateTable(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (dbRepo *CurrencyRepository) doesTableExists() (bool, error) {

	input := &dynamodb.ListTablesInput{}
	for {
		result, err := dbRepo.db.ListTables(context.TODO(), input)
		if err != nil {
			return false, fmt.Errorf("failed to list tables: %w", err)
		}

		for _, tableName := range result.TableNames {
			if tableName == dbRepo.tableName {
				return true, nil
			}
		}

		if result.LastEvaluatedTableName == nil {
			break
		}
		input.ExclusiveStartTableName = result.LastEvaluatedTableName
	}

	return false, nil
}

func (dbRepo *CurrencyRepository) initializeTable() error {
	exists, err := dbRepo.anyItemExists()
	if err != nil {
		return fmt.Errorf("failed to check if any item exists: %w", err)
	}
	if exists {
		return nil
	}

	rates := []models.Rate{
		{Currency: "USD", Rate: 1},
		{Currency: "INR", Rate: 83.25},
		{Currency: "EUR", Rate: 0.92},
	}

	for _, rate := range rates {
		item := map[string]types.AttributeValue{
			"Currency": &types.AttributeValueMemberS{Value: rate.Currency},
			"Rate":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", rate.Rate)},
		}

		input := &dynamodb.PutItemInput{
			TableName: aws.String(dbRepo.tableName),
			Item:      item,
		}

		_, err := dbRepo.db.PutItem(context.TODO(), input)
		if err != nil {
			return fmt.Errorf("failed to insert item (%s -> %f): %w", rate.Currency, rate.Rate, err)
		}
	}

	fmt.Println("Items inserted")
	return nil
}

func (dbRepo *CurrencyRepository) anyItemExists() (bool, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(dbRepo.tableName),
		Limit:     aws.Int32(1), // Limit to 1 item
		Select:    "COUNT",      //Only get the count, not the item data.
	}

	output, err := dbRepo.db.Scan(context.TODO(), input)
	if err != nil {
		return false, err
	}

	return output.Count > 0, nil
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
