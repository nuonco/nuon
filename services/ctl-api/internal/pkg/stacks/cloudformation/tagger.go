package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation/tags"
)

const TagKeyRunnerAssumable = "runner.nuon.co/assumable"

type tagBuilder struct {
	installID  string
	orgID      string
	appID      string
	additional map[string]string
}

func (t tagBuilder) apply(existing []tags.Tag, name string) []tags.Tag {
	existingMap := make(map[string]string)
	for _, tag := range existing {
		existingMap[tag.Key] = tag.Value
	}

	existingMap["install.nuon.co/id"] = t.installID
	existingMap["nuon_install_id"] = t.installID

	// why: Org and app ids are emitted only in the domain-qualified form. The snake_case keys
	// above are legacy duplicates kept for the consumers that already read them; nothing
	// reads a snake_case org or app id off an AWS resource. Skipped when empty so a caller
	// without org/app context emits no tag rather than a blank one, which would silently
	// fail any iam:ResourceTag condition matching on it.
	if t.orgID != "" {
		existingMap["org.nuon.co/id"] = t.orgID
	}
	if t.appID != "" {
		existingMap["app.nuon.co/id"] = t.appID
	}
	if _, has := existingMap[name]; !has {
		if name != "" {
			existingMap["Name"] = fmt.Sprintf("%s-%s", t.installID, name)
		} else {
			existingMap["Name"] = t.installID
		}
	}
	for k, v := range t.additional {
		if _, has := existingMap[k]; !has {
			existingMap[k] = v
		}
	}

	ret := []tags.Tag{}
	for k, v := range existingMap {
		ret = append(ret, tags.Tag{
			Key:   k,
			Value: v,
		})
	}

	return ret
}
