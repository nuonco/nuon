package plan

import (
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// AdmissionReviewInput mimics Kubernetes AdmissionReview structure for OPA policy evaluation.
// This structure matches what existing OPA policies expect (e.g., input.review.object).
type AdmissionReviewInput struct {
	Review AdmissionReviewRequest `json:"review"`
}

// AdmissionReviewRequest contains the object being reviewed and its kind information.
type AdmissionReviewRequest struct {
	Kind   AdmissionReviewKind    `json:"kind"`
	Object map[string]interface{} `json:"object"`
}

// AdmissionReviewKind contains the GVK (Group, Version, Kind) of the object.
type AdmissionReviewKind struct {
	Kind    string `json:"kind"`
	Group   string `json:"group,omitempty"`
	Version string `json:"version,omitempty"`
}

// ParseMultiDocYAMLToAdmissionReviews parses a multi-document YAML string (separated by ---)
// and converts each document into an AdmissionReviewInput structure suitable for OPA policy evaluation.
func ParseMultiDocYAMLToAdmissionReviews(multiDocYAML string) ([]AdmissionReviewInput, error) {
	if strings.TrimSpace(multiDocYAML) == "" {
		return []AdmissionReviewInput{}, nil
	}

	docs := strings.Split(multiDocYAML, "---")
	var results []AdmissionReviewInput

	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var obj map[string]interface{}
		if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal YAML document")
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
