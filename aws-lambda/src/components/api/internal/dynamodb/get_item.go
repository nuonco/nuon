package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/nuonco/example-app-configs/aws-lambda/src/components/api/internal/widget"
)

func GetItem(id, table string) (*widget.Widget, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				N: aws.String(id),
			},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	widget := new(widget.Widget)
	err = dynamodbattribute.UnmarshalMap(result.Item, widget)
	if err != nil {
		return nil, err
	}

	return widget, nil
}
