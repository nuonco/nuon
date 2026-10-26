package cloudformation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	testOrgID   = "orgabcdefghijklmnopqrstuvw"
	testAppID   = "appabcdefghijklmnopqrstuvw"
	testBaseURL = "https://templates.example.com/bucket"
)

var opaqueStackKey = regexp.MustCompile(`^stacks/[0-9a-f]{64}(\.yaml|\.yml|\.json|\.ui\.json)$`)

func TestCustomNestedStackS3Key(t *testing.T) {
	tests := []struct {
		name        string
		templateURL string
		ext         string
	}{
		{name: "yaml extension", templateURL: "https://example.com/template.yaml", ext: ".yaml"},
		{name: "json extension", templateURL: "https://example.com/template.json", ext: ".json"},
		{name: "yml extension", templateURL: "https://example.com/template.yml", ext: ".yml"},
		{name: "no recognized extension defaults to yaml", templateURL: "https://example.com/template", ext: ".yaml"},
		{name: "case insensitive extension", templateURL: "https://example.com/template.JSON", ext: ".json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := CustomNestedStackS3Key(testOrgID, testAppID, "abc123", tt.templateURL)
			assert.Regexp(t, opaqueStackKey, key)
			assert.True(t, strings.HasSuffix(key, tt.ext), key)
		})
	}
}

func TestStackKeysDoNotExposeIDs(t *testing.T) {
	keys := []string{
		CustomNestedStackS3Key(testOrgID, testAppID, "abc123", "vpc.yaml"),
		CustomNestedStackTemplateURL(testBaseURL, testOrgID, testAppID, "abc123", "vpc.yaml"),
		UIDefinitionS3Key(testOrgID, testAppID, "abc123"),
		UIDefinitionSourceURL(testBaseURL, testOrgID, testAppID, "abc123"),
	}
	for _, key := range keys {
		assert.NotContains(t, key, testOrgID)
		assert.NotContains(t, key, testAppID)
		assert.NotContains(t, key, "abc123")
	}
}

func TestCustomNestedStackS3KeyIsScopedToApp(t *testing.T) {
	key := CustomNestedStackS3Key(testOrgID, testAppID, "abc123", "vpc.yaml")
	assert.Equal(t, key, CustomNestedStackS3Key(testOrgID, testAppID, "abc123", "vpc.yaml"))
	assert.NotEqual(t, key, CustomNestedStackS3Key(testOrgID, "appother", "abc123", "vpc.yaml"))
	assert.NotEqual(t, key, CustomNestedStackS3Key("orgother", testAppID, "abc123", "vpc.yaml"))
	assert.NotEqual(t, key, CustomNestedStackS3Key(testOrgID, testAppID, "def456", "vpc.yaml"))
}

func TestCustomNestedStackTemplateURL(t *testing.T) {
	key := CustomNestedStackS3Key(testOrgID, testAppID, "abc123", "vpc.json")
	assert.Equal(t, testBaseURL+"/"+key, CustomNestedStackTemplateURL(testBaseURL, testOrgID, testAppID, "abc123", "vpc.json"))
	assert.Equal(t, testBaseURL+"/"+key, CustomNestedStackTemplateURL(testBaseURL+"/", testOrgID, testAppID, "abc123", "vpc.json"))
}

func TestResolveCustomNestedStackTemplateURL(t *testing.T) {
	t.Run("recorded upload url wins", func(t *testing.T) {
		stack := config.CustomNestedStack{
			TemplateURL:       "./vpc.yaml",
			ContentsHash:      "abc123",
			TemplateSourceURL: "https://elsewhere.example.com/stacks/recorded.yaml",
		}
		assert.Equal(t, stack.TemplateSourceURL, ResolveCustomNestedStackTemplateURL(testBaseURL, testOrgID, testAppID, stack))
	})

	t.Run("uploads without a recorded url resolve to the legacy layout", func(t *testing.T) {
		stack := config.CustomNestedStack{TemplateURL: "./vpc.yaml", ContentsHash: "abc123"}
		assert.Equal(t,
			testBaseURL+"/stacks/"+testOrgID+"/"+testAppID+"/abc123.yaml",
			ResolveCustomNestedStackTemplateURL(testBaseURL, testOrgID, testAppID, stack),
		)
	})

	t.Run("never uploaded falls back to the vendor url", func(t *testing.T) {
		stack := config.CustomNestedStack{TemplateURL: "https://vendor.example.com/vpc.yaml"}
		assert.Equal(t, stack.TemplateURL, ResolveCustomNestedStackTemplateURL(testBaseURL, testOrgID, testAppID, stack))
	})
}

type memoryUploader struct {
	blobs map[string][]byte
}

func (m *memoryUploader) UploadBlob(_ context.Context, blob []byte, key string) error {
	if m.blobs == nil {
		m.blobs = map[string][]byte{}
	}
	m.blobs[key] = blob
	return nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestUploadCustomNestedStackTemplates(t *testing.T) {
	const legacyHash = "1111111111111111111111111111111111111111111111111111111111111111"
	const contents = "Resources: {}\n"

	stackConfig := &app.AppStackConfig{
		OrgID: testOrgID,
		AppID: testAppID,
		CustomNestedStacks: []config.CustomNestedStack{
			{Name: "fresh", TemplateURL: "./fresh.yaml", Contents: contents},
			{Name: "recorded", TemplateURL: "./recorded.yaml", ContentsHash: legacyHash, TemplateSourceURL: "https://old.example.com/stacks/x/y/" + legacyHash + ".yaml"},
			{Name: "legacy", TemplateURL: "./legacy.yaml", ContentsHash: legacyHash},
		},
		UIDefinition: &config.UploadedUIDefinition{Contents: `{"ui":true}`},
	}

	uploader := &memoryUploader{}
	require.NoError(t, UploadCustomNestedStackTemplates(context.Background(), uploader, testBaseURL, stackConfig))

	fresh := stackConfig.CustomNestedStacks[0]
	freshKey := CustomNestedStackS3Key(testOrgID, testAppID, sha256Hex(contents), "./fresh.yaml")
	assert.Regexp(t, opaqueStackKey, freshKey)
	assert.Equal(t, []byte(contents), uploader.blobs[freshKey])
	assert.Equal(t, testBaseURL+"/"+freshKey, fresh.TemplateSourceURL)
	assert.Equal(t, config.CustomNestedStackStatusReady, fresh.Status)
	assert.Empty(t, fresh.Contents)

	assert.Equal(t, "https://old.example.com/stacks/x/y/"+legacyHash+".yaml", stackConfig.CustomNestedStacks[1].TemplateSourceURL)
	assert.Equal(t,
		testBaseURL+"/stacks/"+testOrgID+"/"+testAppID+"/"+legacyHash+".yaml",
		stackConfig.CustomNestedStacks[2].TemplateSourceURL,
	)
	assert.Len(t, uploader.blobs, 2, "already-uploaded stacks must not be re-uploaded")

	ui := stackConfig.UIDefinition
	uiKey := UIDefinitionS3Key(testOrgID, testAppID, sha256Hex(`{"ui":true}`))
	assert.Regexp(t, opaqueStackKey, uiKey)
	assert.Equal(t, []byte(`{"ui":true}`), uploader.blobs[uiKey])
	assert.Equal(t, testBaseURL+"/"+uiKey, ui.SourceURL)

	for key := range uploader.blobs {
		assert.NotContains(t, key, testOrgID)
		assert.NotContains(t, key, testAppID)
	}
}

func TestUploadUIDefinitionBackfill(t *testing.T) {
	t.Run("keeps the recorded url", func(t *testing.T) {
		ui := &config.UploadedUIDefinition{ContentsHash: "abc123", SourceURL: "https://old.example.com/ui.json"}
		require.NoError(t, UploadUIDefinition(context.Background(), &memoryUploader{}, testBaseURL, testOrgID, testAppID, ui))
		assert.Equal(t, "https://old.example.com/ui.json", ui.SourceURL)
		assert.Equal(t, config.CustomNestedStackStatusReady, ui.Status)
	})

	t.Run("points unrecorded uploads at the legacy layout", func(t *testing.T) {
		ui := &config.UploadedUIDefinition{ContentsHash: "abc123"}
		require.NoError(t, UploadUIDefinition(context.Background(), &memoryUploader{}, testBaseURL, testOrgID, testAppID, ui))
		assert.Equal(t, testBaseURL+"/stacks/"+testOrgID+"/"+testAppID+"/abc123.ui.json", ui.SourceURL)
	})
}
