package dynamodb

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type Client struct {
	svc       *dynamodb.Client
	tableName string
}

func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS configuration", err)
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME environment variable not found")
	}

	return &Client{
		svc:       dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil

}

func (c *Client) SaveForm(ctx context.Context, name, email, message, targetChannel string) error {
	recordID := uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339)

	item := map[string]types.AttributeValue{
		"id":             &types.AttributeValueMemberS{Value: recordID},
		"created_at":     &types.AttributeValueMemberS{Value: timestamp},
		"name":           &types.AttributeValueMemberS{Value: name},
		"email":          &types.AttributeValueMemberS{Value: email},
		"message":        &types.AttributeValueMemberS{Value: message},
		"target_channel": &types.AttributeValueMemberS{Value: targetChannel},
	}

	if _, err := c.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &c.tableName,
		Item:      item,
	}); err != nil {
		return fmt.Errorf("Error saving record to DynamoDB: %w", err)
	}
	return nil
}
