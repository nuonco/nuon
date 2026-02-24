package dynamodb

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	region = os.Getenv("AWS_REGION")
	db     = dynamodb.New(session.New(), aws.NewConfig().WithRegion(region))
)
