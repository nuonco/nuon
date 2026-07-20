package app

import (
	"context"
	"encoding/base64"
	"testing"

	"gorm.io/gorm"
)

func TestAWSAccountConnectionBeforeCreate(t *testing.T) {
	connection := &AWSAccountConnection{}
	if err := connection.BeforeCreate(&gorm.DB{Statement: &gorm.Statement{Context: context.Background()}}); err != nil {
		t.Fatal(err)
	}
	if len(connection.ID) != 26 || connection.VerificationStatus != AWSAccountConnectionVerificationPending {
		t.Fatalf("unexpected defaults: %#v", connection)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(connection.ExternalID)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("external ID is not 32 random bytes: %q", connection.ExternalID)
	}
}

func TestAWSAccountConnectionsFeatureIsAdminManaged(t *testing.T) {
	features := GetFeatures()
	if features[len(features)-1] != OrgFeatureAWSAccountConnections {
		t.Fatal("AWS account connections must remain the newest feature")
	}
	for _, feature := range GetUserManageableFeatures() {
		if feature == OrgFeatureAWSAccountConnections {
			t.Fatal("AWS account connections must not be user manageable")
		}
	}
}
