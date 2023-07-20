package client

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultOrgServer(t *testing.T) {
	orgID := uuid.NewString()
	rootDomain := "test.nuon.co"

	addr := DefaultOrgServerAddress(rootDomain, orgID)
	assert.Equal(t, fmt.Sprintf("%s.%s:9701", orgID, rootDomain), addr)
}

func TestDefaultOrgDomain(t *testing.T) {
	orgID := uuid.NewString()
	rootDomain := "test.nuon.co"

	domain := DefaultOrgServerDomain(rootDomain, orgID)
	assert.Equal(t, fmt.Sprintf("%s.%s", orgID, rootDomain), domain)
}
