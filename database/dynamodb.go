package database

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
)

type DynamoDB struct {
	DB        *dynamodb.Client
	TableName string
}

func InitDynamoDB() (*DynamoDB, error) {
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
	newDbInstance := &DynamoDB{
		DB:        dynamodb.NewFromConfig(awsConfig),
		TableName: "conversions",
	}
	if err := newDbInstance.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	if err := newDbInstance.initializeTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}
	return newDbInstance, nil
}

func (d *DynamoDB) createTable() error {
	// Check if the table already exists
	exists, err := d.doesTableExists()
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if exists {
		// Table already exists, no need to create
		return nil
	}

	// Define the table schema
	input := &dynamodb.CreateTableInput{
		TableName: &d.TableName,
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
	_, err = d.DB.CreateTable(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func (d *DynamoDB) doesTableExists() (bool, error) {

	input := &dynamodb.ListTablesInput{}
	for {
		result, err := d.DB.ListTables(context.TODO(), input)
		if err != nil {
			return false, fmt.Errorf("failed to list tables: %w", err)
		}

		for _, tableName := range result.TableNames {
			if tableName == d.TableName {
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

func (d *DynamoDB) initializeTable() error {
	exists, err := d.anyItemExists()
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
			TableName: aws.String(d.TableName),
			Item:      item,
		}

		_, err := d.DB.PutItem(context.TODO(), input)
		if err != nil {
			return fmt.Errorf("failed to insert item (%s -> %f): %w", rate.Currency, rate.Rate, err)
		}
	}

	fmt.Println("Items inserted")
	return nil
}

func (d *DynamoDB) anyItemExists() (bool, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(d.TableName),
		Limit:     aws.Int32(1), // Limit to 1 item
		Select:    "COUNT",      //Only get the count, not the item data.
	}

	output, err := d.DB.Scan(context.TODO(), input)
	if err != nil {
		return false, err
	}

	return output.Count > 0, nil
}
