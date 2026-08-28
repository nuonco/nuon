package customermanaged

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	InstallationRegistrationSchemaVersion = 2
	InstallationRegistrationKey           = "registration/installation.json"
)

var (
	registrationDigestPattern     = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	registrationDeploymentPattern = regexp.MustCompile(`^[a-z0-9]{1,8}$`)
	registrationProviderPattern   = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
)

type InstallationRegistration struct {
	SchemaVersion  int                           `json:"schema_version"`
	RegistrationID string                        `json:"registration_id"`
	ReleaseID      string                        `json:"release_id"`
	ReleaseDigest  string                        `json:"release_digest"`
	PackageID      string                        `json:"package_id"`
	PackageDigest  string                        `json:"package_digest"`
	BundleDigest   string                        `json:"bundle_digest"`
	ArchiveDigest  string                        `json:"archive_digest"`
	OperationID    string                        `json:"operation_id"`
	DeploymentID   string                        `json:"deployment_id"`
	InstallID      string                        `json:"install_id"`
	Cloud          InstallationRegistrationCloud `json:"cloud"`
	Stack          InstallationRegistrationStack `json:"stack"`
	InstalledAt    time.Time                     `json:"installed_at"`
}

type InstallationRegistrationCloud struct {
	Provider  string `json:"provider"`
	AccountID string `json:"account_id"`
	Region    string `json:"region"`
}

type InstallationRegistrationStack struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewInstallationRegistration(registration InstallationRegistration) (InstallationRegistration, error) {
	registration = normalizeInstallationRegistration(registration)
	registration.RegistrationID = ""
	if err := validateInstallationRegistrationFields(registration); err != nil {
		return InstallationRegistration{}, err
	}
	raw, err := json.Marshal(registration)
	if err != nil {
		return InstallationRegistration{}, fmt.Errorf("encode installation registration: %w", err)
	}
	digest := sha256.Sum256(raw)
	registration.RegistrationID = "airreg_" + hex.EncodeToString(digest[:])[:24]
	return registration, nil
}

func (r InstallationRegistration) Validate() error {
	wanted := strings.TrimSpace(r.RegistrationID)
	canonical, err := NewInstallationRegistration(r)
	if err != nil {
		return err
	}
	if wanted != canonical.RegistrationID {
		return fmt.Errorf("installation registration ID does not match its contents")
	}
	return nil
}

func normalizeInstallationRegistration(registration InstallationRegistration) InstallationRegistration {
	registration.RegistrationID = strings.TrimSpace(registration.RegistrationID)
	registration.ReleaseID = strings.TrimSpace(registration.ReleaseID)
	registration.ReleaseDigest = strings.ToLower(strings.TrimSpace(registration.ReleaseDigest))
	registration.PackageID = strings.TrimSpace(registration.PackageID)
	registration.PackageDigest = strings.ToLower(strings.TrimSpace(registration.PackageDigest))
	registration.BundleDigest = strings.ToLower(strings.TrimSpace(registration.BundleDigest))
	registration.ArchiveDigest = strings.ToLower(strings.TrimSpace(registration.ArchiveDigest))
	registration.OperationID = strings.TrimSpace(registration.OperationID)
	registration.DeploymentID = strings.TrimSpace(registration.DeploymentID)
	registration.InstallID = strings.TrimSpace(registration.InstallID)
	registration.Cloud.Provider = strings.ToLower(strings.TrimSpace(registration.Cloud.Provider))
	registration.Cloud.AccountID = strings.TrimSpace(registration.Cloud.AccountID)
	registration.Cloud.Region = strings.TrimSpace(registration.Cloud.Region)
	registration.Stack.Type = strings.ToLower(strings.TrimSpace(registration.Stack.Type))
	registration.Stack.ID = strings.TrimSpace(registration.Stack.ID)
	registration.Stack.Name = strings.TrimSpace(registration.Stack.Name)
	registration.InstalledAt = registration.InstalledAt.UTC()
	return registration
}

func validateInstallationRegistrationFields(registration InstallationRegistration) error {
	if registration.SchemaVersion != 1 && registration.SchemaVersion != InstallationRegistrationSchemaVersion {
		return fmt.Errorf("unsupported installation registration schema version %d", registration.SchemaVersion)
	}
	if registration.ReleaseID == "" || len(registration.ReleaseID) > 255 {
		return fmt.Errorf("release ID must be between 1 and 255 characters")
	}
	if !registrationDigestPattern.MatchString(registration.ReleaseDigest) {
		return fmt.Errorf("release digest must be a sha256 digest")
	}
	if registration.PackageID == "" || len(registration.PackageID) > 255 {
		return fmt.Errorf("package ID must be between 1 and 255 characters")
	}
	if !registrationDigestPattern.MatchString(registration.PackageDigest) {
		return fmt.Errorf("package digest must be a sha256 digest")
	}
	if !registrationDigestPattern.MatchString(registration.BundleDigest) {
		return fmt.Errorf("bundle digest must be a sha256 digest")
	}
	if !registrationDigestPattern.MatchString(registration.ArchiveDigest) {
		return fmt.Errorf("archive digest must be a sha256 digest")
	}
	if registration.SchemaVersion >= 2 && (registration.OperationID == "" || len(registration.OperationID) > 255) {
		return fmt.Errorf("operation ID must be between 1 and 255 characters")
	}
	if !registrationDeploymentPattern.MatchString(registration.DeploymentID) {
		return fmt.Errorf("deployment ID must be 1-8 lowercase letters or digits")
	}
	if registration.InstallID == "" || len(registration.InstallID) > 255 {
		return fmt.Errorf("install ID must be between 1 and 255 characters")
	}
	if !registrationProviderPattern.MatchString(registration.Cloud.Provider) {
		return fmt.Errorf("cloud provider is invalid")
	}
	if registration.Cloud.AccountID == "" || len(registration.Cloud.AccountID) > 255 {
		return fmt.Errorf("cloud account ID must be between 1 and 255 characters")
	}
	if registration.Cloud.Region == "" || len(registration.Cloud.Region) > 255 {
		return fmt.Errorf("cloud region must be between 1 and 255 characters")
	}
	if registration.Stack.Type == "" || len(registration.Stack.Type) > 64 {
		return fmt.Errorf("stack type must be between 1 and 64 characters")
	}
	if registration.Stack.ID == "" || len(registration.Stack.ID) > 2048 {
		return fmt.Errorf("stack ID must be between 1 and 2048 characters")
	}
	if registration.Stack.Name == "" || len(registration.Stack.Name) > 255 {
		return fmt.Errorf("stack name must be between 1 and 255 characters")
	}
	if registration.InstalledAt.IsZero() {
		return fmt.Errorf("installation completion time is required")
	}
	return nil
}
