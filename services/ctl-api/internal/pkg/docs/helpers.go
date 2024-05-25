package docs

import (
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

// addSpecTags will add tags for each operation into the top level, general section
func addSpecTags(doc *openapi2.T) {
	allTags := make(map[string]struct{}, 0)
	for _, pathItem := range doc.Paths {
		for _, op := range pathItem.Operations() {
			for _, tag := range op.Tags {
				allTags[tag] = struct{}{}
			}
		}
	}

	doc.Tags = make([]*openapi3.Tag, 0, len(allTags))
	for tag := range allTags {
		doc.Tags = append(doc.Tags, &openapi3.Tag{
			Name:        tag,
			Description: tag,
		})
	}
}

// removeSecurity removes the security params
func removeSecurity(doc *openapi2.T) {
	doc.SecurityDefinitions = nil
}
