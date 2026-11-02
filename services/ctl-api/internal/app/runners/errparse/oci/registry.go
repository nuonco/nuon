package oci

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type registryError struct {
	Kind    compositeerrors.Type `json:"kind"`
	Message string               `json:"message"`
	Help    string               `json:"help"`
	Output  string               `json:"output"`
}

func (e *registryError) Error() string                      { return e.Message }
func (e *registryError) Type() compositeerrors.Type         { return e.Kind }
func (e *registryError) Severity() compositeerrors.Severity { return compositeerrors.SeverityError }
func (e *registryError) Sections() []compositeerrors.Section {
	return []compositeerrors.Section{
		compositeerrors.TextSection("How to resolve", e.Help),
		compositeerrors.CodeSection("Error output", e.Output),
	}
}

func parseRegistry(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	if ctx.Tool != errparse.ToolOCI {
		return nil
	}
	raw := ctx.Raw
	e := &registryError{}
	switch {
	case strings.Contains(raw, "/v2/") && (strings.Contains(raw, "response status code 401:") || strings.Contains(raw, "response status code 403:")):
		e.Kind = "oci.registry-access"
		e.Message = "Image registry rejected access"
		e.Help = "Check the credentials and required read/write permissions for the registry named in Error output. For source images, review the pull identity selected in Settings > Image repositories and its permissions on the exact repository. A 401 can also mean invalid or expired credentials. Registries may hide missing repositories behind access errors."
		switch {
		case strings.Contains(raw, ".dkr.ecr."):
			e.Help += " ECR pulls require ecr:GetAuthorizationToken, ecr:BatchGetImage, and ecr:GetDownloadUrlForLayer."
		case strings.Contains(raw, "-docker.pkg.dev/"):
			e.Help += " GAR pulls require Artifact Registry Reader on this repository (artifactregistry.repositories.downloadArtifacts)."
		case strings.Contains(raw, ".azurecr.io/"):
			e.Help += " ACR pulls require Container Registry Repository Reader and its repository condition on ABAC registries, or AcrPull on RBAC-only registries."
		}
	case strings.Contains(raw, "unable to resolve") && (strings.Contains(raw, ": not found") || strings.Contains(raw, "response status code 404:")):
		e.Kind = "oci.image-not-found"
		e.Message = "Image reference not found"
		e.Help = "Check the registry, repository/image path, and exact tag or digest in the component configuration. Publish the image or correct the reference, then retry. The registry response does not always distinguish a missing repository from a missing tag."
	default:
		return nil
	}
	output := []rune(strings.TrimSpace(raw))
	if len(output) > 8000 {
		output = append(output[:8000], '…')
	}
	e.Output = string(output)
	return e
}

func init() {
	errparse.Register(errparse.NewParser(errparse.LayerToolSpecific, parseRegistry,
		errparse.WithTools(errparse.ToolOCI),
		errparse.WithSignals("response status code 401:", "response status code 403:", "response status code 404:", ": not found"),
	))
}
