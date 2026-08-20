package plan

import (
	"io"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// AdmissionReviewInput mimics Kubernetes AdmissionReview structure for OPA policy evaluation.
// This structure matches what existing OPA policies expect (e.g., input.review.object).
type AdmissionReviewInput struct {
	Review AdmissionReviewRequest `json:"review"`
}

// AdmissionReviewRequest contains the object being reviewed and its kind information.
type AdmissionReviewRequest struct {
	Kind AdmissionReviewKind `json:"kind"`
	// Operation is the plan operation: CREATE, UPDATE or DELETE. Omitted when unknown.
	Operation string                 `json:"operation,omitempty"`
	Object    map[string]interface{} `json:"object"`
}

// AdmissionReviewKind contains the GVK (Group, Version, Kind) of the object.
type AdmissionReviewKind struct {
	Kind    string `json:"kind"`
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
}

// ParseMultiDocYAMLToAdmissionReviews parses a multi-document YAML stream
// and converts each document into an AdmissionReviewInput structure suitable for OPA policy evaluation.
func ParseMultiDocYAMLToAdmissionReviews(multiDocYAML string) ([]AdmissionReviewInput, error) {
	if strings.TrimSpace(multiDocYAML) == "" {
		return []AdmissionReviewInput{}, nil
	}

	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(multiDocYAML), 4096)
	results := make([]AdmissionReviewInput, 0)
	for documentIndex := 1; ; documentIndex++ {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrapf(err, "failed to unmarshal YAML document %d", documentIndex)
		}

		if len(obj) == 0 {
			continue
		}

		reviewInput := AdmissionReviewInput{
			Review: AdmissionReviewRequest{
				Kind:   extractKindInfo(obj),
				Object: obj,
			},
		}

		results = append(results, reviewInput)
	}

	return results, nil
}

// extractKindInfo extracts the GVK information from a Kubernetes object.
func extractKindInfo(obj map[string]interface{}) AdmissionReviewKind {
	kind := AdmissionReviewKind{}

	if k, ok := obj["kind"].(string); ok {
		kind.Kind = k
	}

	if apiVersion, ok := obj["apiVersion"].(string); ok {
		parts := strings.Split(apiVersion, "/")
		if len(parts) == 2 {
			kind.Group = parts[0]
			kind.Version = parts[1]
		} else if len(parts) == 1 {
			kind.Version = parts[0]
		}
	}

	return kind
}
