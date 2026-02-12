package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func PutItem(id, table string) error {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				N: aws.String(id),
			},
		},
	}

	_, err := db.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}
