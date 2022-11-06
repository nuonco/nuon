package provision

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/stretchr/testify/assert"
)

func TestInstallationPrefix(t *testing.T) {
	orgID := uuid.NewString()
	orgShortID, err := shortid.ParseString(orgID)
	assert.Nil(t, err)

	appID := uuid.NewString()
	appShortID, err := shortid.ParseString(appID)
	assert.Nil(t, err)

	installID := uuid.NewString()
	installShortID, err := shortid.ParseString(installID)
	assert.Nil(t, err)

	prefix := getInstallationPrefix(orgShortID, appShortID, installShortID)

	expected := fmt.Sprintf("installations/org=%s/app=%s/install=%s", orgShortID, appShortID, installShortID)
	assert.Equal(t, expected, prefix)
}
