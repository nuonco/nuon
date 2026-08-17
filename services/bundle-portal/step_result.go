package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nuonco/nuon/pkg/plans"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/statestore"
)

// stepResult serves the execution result a deploy handler persisted for a
// bootstrap step, with the compressed payload decoded so the portal can render
// the real artifact: the `terraform plan` JSON for terraform steps, or the
// computed resource diffs for helm and kubernetes manifest steps.
func (p *portalServer) stepResult(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !jobIDPattern.MatchString(id) {
		writeAPIError(w, fmt.Errorf("invalid step ID"), http.StatusBadRequest)
		return
	}
	raw, ok, err := p.store.Get(r.Context(), statestore.StepResultKey(id))
	if err == nil && !ok {
		if runID, stepID, found := strings.Cut(id, "--"); found && runID != "" && stepID != "" {
			raw, ok, err = p.store.Get(r.Context(), day2.RunStepResultKey(runID, stepID))
		}
	}
	if err != nil {
		writeAPIError(w, err, http.StatusInternalServerError)
		return
	}
	if !ok {
		writeAPIError(w, fmt.Errorf("not found"), http.StatusNotFound)
		return
	}
	var result struct {
		Success                   bool            `json:"success"`
		Contents                  string          `json:"contents"`
		ContentsCompressed        string          `json:"contents_compressed"`
		ContentsDisplay           json.RawMessage `json:"contents_display"`
		ContentsDisplayCompressed string          `json:"contents_display_compressed"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		writeAPIError(w, fmt.Errorf("decode step result %s: %w", id, err), http.StatusInternalServerError)
		return
	}
	content, err := decodeResultContent(result.ContentsDisplayCompressed, result.ContentsCompressed, result.ContentsDisplay, result.Contents)
	if err != nil {
		writeAPIError(w, fmt.Errorf("decode step result %s contents: %w", id, err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"success": result.Success,
		"kind":    classifyResultContent(content),
		"content": content,
	})
}

// decodeResultContent picks the richest representation the runner uploaded:
// display-compressed and compressed payloads are gzip+base64 (terraform
// plan.json, helm/k8s plan contents), the rest are used as-is.
func decodeResultContent(displayCompressed, compressed string, display json.RawMessage, plain string) (json.RawMessage, error) {
	for _, encoded := range []string{displayCompressed, compressed} {
		if encoded == "" {
			continue
		}
		decoded, err := plans.DecompressPlan(encoded)
		if err != nil {
			return nil, err
		}
		return asRawJSON(decoded), nil
	}
	if len(display) > 0 && string(display) != "null" {
		return display, nil
	}
	if plain != "" {
		return asRawJSON([]byte(plain)), nil
	}
	return json.RawMessage("null"), nil
}

func asRawJSON(data []byte) json.RawMessage {
	if json.Valid(data) {
		return data
	}
	quoted, err := json.Marshal(string(data))
	if err != nil {
		return json.RawMessage("null")
	}
	return quoted
}

func classifyResultContent(content json.RawMessage) string {
	var probe struct {
		TerraformVersion string            `json:"terraform_version"`
		ResourceChanges  []json.RawMessage `json:"resource_changes"`
		PlannedValues    json.RawMessage   `json:"planned_values"`
		HelmContentDiff  []json.RawMessage `json:"helm_content_diff"`
		K8sContentDiff   []json.RawMessage `json:"k8s_content_diff"`
	}
	if err := json.Unmarshal(content, &probe); err != nil {
		return "unknown"
	}
	switch {
	case probe.HelmContentDiff != nil:
		return "helm"
	case probe.K8sContentDiff != nil:
		return "kubernetes_manifest"
	case probe.TerraformVersion != "" || probe.ResourceChanges != nil || len(probe.PlannedValues) > 0:
		return "terraform"
	}
	return "unknown"
}
