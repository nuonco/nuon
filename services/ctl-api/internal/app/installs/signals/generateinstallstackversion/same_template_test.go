package generateinstallstackversion

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSameTemplate(t *testing.T) {
	active := &app.InstallStackVersion{ID: "istactive", PhoneHomeID: "phactive"}
	next := &app.InstallStackVersion{ID: "istnext", PhoneHomeID: "phnext"}

	tmpl := func(phoneHomeID, versionID, bucket string) []byte {
		return []byte(`{"Resources":{"PhoneHome":{"Properties":{"Url":"https://x/phone-home/` + phoneHomeID + `","Key":"` + versionID + `/stack.json"}},"Bucket":{"Name":"` + bucket + `"}},"AWSTemplateFormatVersion":"2010-09-09"}`)
	}
	// jsonb reorders keys and drops whitespace
	stored := func(phoneHomeID, versionID, bucket string) []byte {
		return []byte(`{"AWSTemplateFormatVersion": "2010-09-09", "Resources": {"Bucket": {"Name": "` + bucket + `"}, "PhoneHome": {"Properties": {"Key": "` + versionID + `/stack.json", "Url": "https://x/phone-home/` + phoneHomeID + `"}}}}`)
	}

	assert.True(t, SameTemplate(tmpl("phnext", "istnext", "b"), next, stored("phactive", "istactive", "b"), active))
	assert.False(t, SameTemplate(tmpl("phnext", "istnext", "changed"), next, stored("phactive", "istactive", "b"), active))
	assert.False(t, SameTemplate(tmpl("phnext", "istnext", "b"), next, nil, active))
	assert.False(t, SameTemplate([]byte("not json"), next, stored("phactive", "istactive", "b"), active))
	assert.False(t, SameTemplate(tmpl("phnext", "istnext", "b"), next, stored("phactive", "istactive", "b"), nil))
}

func TestSameTemplateIgnoresTagAndParameterGroupOrder(t *testing.T) {
	a := &app.InstallStackVersion{ID: "ista"}
	b := &app.InstallStackVersion{ID: "istb"}
	sorted := []byte(`{"Metadata":{"AWS::CloudFormation::Interface":{"ParameterGroups":[{"Label":{"default":"Runner"},"Parameters":["A","B"]}]}},"Resources":{"R":{"Properties":{"Tags":[{"Key":"a","Value":"1"},{"Key":"b","Value":"2"}],"Name":{"Fn::Join":["-",["x","y"]]}}}}}`)
	shuffled := []byte(`{"Metadata":{"AWS::CloudFormation::Interface":{"ParameterGroups":[{"Label":{"default":"Runner"},"Parameters":["B","A"]}]}},"Resources":{"R":{"Properties":{"Tags":[{"Key":"b","Value":"2"},{"Key":"a","Value":"1"}],"Name":{"Fn::Join":["-",["x","y"]]}}}}}`)
	joinSwapped := []byte(`{"Metadata":{"AWS::CloudFormation::Interface":{"ParameterGroups":[{"Label":{"default":"Runner"},"Parameters":["A","B"]}]}},"Resources":{"R":{"Properties":{"Tags":[{"Key":"a","Value":"1"},{"Key":"b","Value":"2"}],"Name":{"Fn::Join":["-",["y","x"]]}}}}}`)

	assert.True(t, SameTemplate(sorted, a, shuffled, b))
	assert.False(t, SameTemplate(sorted, a, joinSwapped, b))
}
