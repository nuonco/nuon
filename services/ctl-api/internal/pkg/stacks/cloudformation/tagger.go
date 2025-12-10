package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation/tags"
)

type tagBuilder struct {
	installID  string
	additional map[string]string
}

func (t tagBuilder) apply(existing []tags.Tag, name string) []tags.Tag {
	existingMap := make(map[string]string)
	for _, tag := range existing {
		existingMap[tag.Key] = tag.Value
	}

	existingMap["install.nuon.co/id"] = t.installID
	existingMap["nuon_install_id"] = t.installID
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
