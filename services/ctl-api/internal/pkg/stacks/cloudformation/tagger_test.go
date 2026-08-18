package cloudformation

import (
	"testing"

	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/stretchr/testify/assert"
)

func tagMap(applied []tags.Tag) map[string]string {
	out := make(map[string]string, len(applied))
	for _, tag := range applied {
		out[tag.Key] = tag.Value
	}

	return out
}

func TestTagBuilderApply_EntityTags(t *testing.T) {
	tb := tagBuilder{installID: "inl123", orgID: "org123", appID: "app123"}

	applied := tagMap(tb.apply(nil, "vpc"))

	assert.Equal(t, "inl123", applied["install.nuon.co/id"])
	assert.Equal(t, "inl123", applied["nuon_install_id"])
	assert.Equal(t, "org123", applied["org.nuon.co/id"])
	assert.Equal(t, "app123", applied["app.nuon.co/id"])
	assert.Equal(t, "inl123-vpc", applied["Name"])
}

// A blank id would silently fail any iam:ResourceTag condition matching on it, so an
// absent org/app emits no tag at all.
func TestTagBuilderApply_OmitsBlankEntityTags(t *testing.T) {
	tb := tagBuilder{installID: "inl123"}

	applied := tagMap(tb.apply(nil, "vpc"))

	assert.NotContains(t, applied, "org.nuon.co/id")
	assert.NotContains(t, applied, "app.nuon.co/id")
}

func TestTagBuilderApply_CustomerTagsCannotOverrideEntityTags(t *testing.T) {
	tb := tagBuilder{
		installID: "inl123",
		orgID:     "org123",
		appID:     "app123",
		additional: map[string]string{
			"install.nuon.co/id": "spoofed",
			"org.nuon.co/id":     "spoofed",
			"app.nuon.co/id":     "spoofed",
			"team":               "platform",
		},
	}

	applied := tagMap(tb.apply(nil, "vpc"))

	assert.Equal(t, "inl123", applied["install.nuon.co/id"])
	assert.Equal(t, "org123", applied["org.nuon.co/id"])
	assert.Equal(t, "app123", applied["app.nuon.co/id"])
	assert.Equal(t, "platform", applied["team"], "unreserved customer tags still pass through")
}
