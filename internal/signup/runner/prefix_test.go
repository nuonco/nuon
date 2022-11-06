package runner

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/stretchr/testify/assert"
)

func TestOrgPrefix(t *testing.T) {
	orgID := uuid.NewString()
	orgShortID, err := shortid.ParseString(orgID)
	assert.Nil(t, err)

	appID := uuid.NewString()
	appShortID, err := shortid.ParseString(appID)
	assert.Nil(t, err)

	prefix := getOrgPrefix(orgShortID, appShortID)

	expected := fmt.Sprintf("installations/org=%s/app=%s", orgShortID, appShortID)
	assert.Equal(t, expected, prefix)
}
