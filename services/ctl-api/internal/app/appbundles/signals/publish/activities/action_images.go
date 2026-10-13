package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/distribution/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	appbundle "github.com/nuonco/nuon/pkg/appbundle"
	bundle "github.com/nuonco/nuon/pkg/appbundle/bundle"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci"
)

func bundleActionImages(ctx context.Context, envelope *appbundle.Envelope, logical *bundle.LogicalManifest, roots *[]bundle.Root, buildIDs map[string]string) error {
	for i := range envelope.Actions {
		action := &envelope.Actions[i]
		var composite plantypes.CompositePlan
		if err := json.Unmarshal(action.CompositePlan, &composite); err != nil {
			return fmt.Errorf("decode action %s: %w", action.ID, err)
		}
		plan := composite.ActionWorkflowRunPlan
		if plan == nil || plan.SourceImage == "" {
			continue
		}
		if plan.ImageRegistry == nil {
			return fmt.Errorf("action %s has no destination image registry", action.ID)
		}
		root, repository, err := actionImageRoot(ctx, plan.SourceImage, *logical, *roots)
		if err != nil {
			return fmt.Errorf("package image for action %s: %w", action.ID, err)
		}
		name := action.Name
		if name == "" {
			name = action.ID
		}
		imageName := "actions/" + name
		logical.Images = append(logical.Images, bundle.Image{Name: imageName, Repository: repository, Artifact: bundleArtifact(root.Descriptor)})
		*roots = append(*roots, root)
		buildIDs["image:"+imageName] = root.Descriptor.Digest.String()

		plan.ImageTag = "action-" + root.Descriptor.Digest.Encoded()
		plan.ImageDigestRef = plan.ImageRegistry.Repository + "@" + root.Descriptor.Digest.String()
		plan.SourceImage = plan.ImageDigestRef
		action.CompositePlan, err = json.Marshal(&composite)
		if err != nil {
			return fmt.Errorf("encode action %s: %w", action.ID, err)
		}
		syncPlan, err := json.Marshal(&plantypes.CompositePlan{SyncOCIPlan: &plantypes.SyncOCIPlan{
			Src: &configs.OCIRegistryRepository{
				RegistryType: configs.OCIRegistryTypePublicOCI,
				Repository:   repository,
			},
			SrcTag: root.Descriptor.Digest.String(), Dst: plan.ImageRegistry, DstTag: plan.ImageTag,
		}})
		if err != nil {
			return fmt.Errorf("encode image sync for action %s: %w", action.ID, err)
		}
		envelope.Steps = append(envelope.Steps, appbundle.Step{
			ID: "sync-action-image-" + action.ID, Name: name + " action image sync",
			JobType: "oci-sync", JobGroup: "sync", JobOperation: "exec",
			DependsOn: []string{"sandbox-apply"}, CompositePlan: syncPlan,
		})
	}
	*roots = uniqueRoots(*roots)
	return envelope.Validate()
}

func actionImageRoot(ctx context.Context, imageRef string, logical bundle.LogicalManifest, roots []bundle.Root) (bundle.Root, string, error) {
	for _, image := range logical.Images {
		if imageRef != appbundle.ComponentOutputPlaceholder(image.Name, "image.ref") {
			continue
		}
		for _, root := range roots {
			if root.Descriptor.Digest.String() != image.Artifact.Digest {
				continue
			}
			desc, err := selectPlatformImage(ctx, root.Source, root.Descriptor, logical.Target)
			if err != nil {
				return bundle.Root{}, "", err
			}
			return bundle.Root{Descriptor: desc, Source: root.Source}, image.Repository, nil
		}
		return bundle.Root{}, "", fmt.Errorf("component image %s is not packaged", image.Name)
	}
	if strings.Contains(imageRef, "__NUON_") {
		return bundle.Root{}, "", fmt.Errorf("image must be a public reference or a bundled container_image component's image.ref output; runtime-dependent image references cannot be packaged")
	}
	named, err := reference.ParseDockerRef(imageRef)
	if err != nil {
		return bundle.Root{}, "", fmt.Errorf("invalid image reference %q: %w", imageRef, err)
	}
	ref := "latest"
	if pinned, ok := named.(reference.Digested); ok {
		ref = pinned.Digest().String()
	} else if tagged, ok := named.(reference.Tagged); ok {
		ref = tagged.Tag()
	}
	repo, err := oci.GetRepo(ctx, &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI, Repository: named.Name(), LoginServer: reference.Domain(named),
	})
	if err != nil {
		return bundle.Root{}, "", err
	}
	desc, err := resolvePlatformImage(ctx, repo, ref, logical.Target)
	if err != nil {
		return bundle.Root{}, "", fmt.Errorf("resolve %s: %w", imageRef, err)
	}
	return bundle.Root{Descriptor: desc, Source: repo}, named.Name(), nil
}

func selectPlatformImage(ctx context.Context, store oras.ReadOnlyTarget, desc ocispec.Descriptor, target bundle.Target) (ocispec.Descriptor, error) {
	if desc.Size > maxFetchedAsset {
		return ocispec.Descriptor{}, fmt.Errorf("image metadata exceeds %d bytes", maxFetchedAsset)
	}
	body, err := content.FetchAll(ctx, store, desc)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if desc.MediaType == ocispec.MediaTypeImageIndex || desc.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" {
		var index ocispec.Index
		if err := json.Unmarshal(body, &index); err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("decode image index: %w", err)
		}
		for _, manifest := range index.Manifests {
			if manifest.Platform != nil && manifest.Platform.OS == target.OS && manifest.Platform.Architecture == target.Architecture &&
				(manifest.MediaType == ocispec.MediaTypeImageManifest || manifest.MediaType == "application/vnd.docker.distribution.manifest.v2+json") {
				return selectPlatformImage(ctx, store, manifest, target)
			}
		}
		return ocispec.Descriptor{}, fmt.Errorf("image index has no %s/%s manifest", target.OS, target.Architecture)
	}
	if desc.MediaType != ocispec.MediaTypeImageManifest && desc.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		return ocispec.Descriptor{}, fmt.Errorf("unsupported image manifest type %q", desc.MediaType)
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("decode image manifest: %w", err)
	}
	if manifest.Config.Size > maxFetchedAsset {
		return ocispec.Descriptor{}, fmt.Errorf("image config exceeds %d bytes", maxFetchedAsset)
	}
	body, err = content.FetchAll(ctx, store, manifest.Config)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("fetch image config: %w", err)
	}
	var image ocispec.Image
	if err := json.Unmarshal(body, &image); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("decode image config: %w", err)
	}
	if image.OS != target.OS || image.Architecture != target.Architecture {
		return ocispec.Descriptor{}, fmt.Errorf("image platform %s/%s does not match bundle target %s/%s", image.OS, image.Architecture, target.OS, target.Architecture)
	}
	desc.Platform = nil
	return desc, nil
}
