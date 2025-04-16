package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateInstallStackVersionRequest struct {
	InstallID      string `validate:"required"`
	InstallStackID string `validate:"required"`
	AppConfigID    string `validate:"required"`
	Region         string `validate:"required"`
	StackName      string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateInstallStackVersion(ctx context.Context, req *CreateInstallStackVersionRequest) (*app.InstallStackVersion, error) {
	phoneHomeID := domains.NewAWSAccountID()
	id := domains.NewInstallStackID()
	bucketKey := fmt.Sprintf("templates/%s/%s.json", req.InstallID, id)
	templateURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(a.cfg.AWSCloudFormationStackTemplateBaseURL, "/"), bucketKey)
	quickLinkURL := fmt.Sprintf("https://%s.console.aws.amazon.com/cloudformation/home?region=%s#/stacks/quickcreate?templateUrl=%s&stackName=%s",
		req.Region, req.Region, templateURL, req.StackName,
	)

	obj := app.InstallStackVersion{
		ID:             id,
		AppConfigID:    req.AppConfigID,
		InstallID:      req.InstallID,
		InstallStackID: req.InstallStackID,
		PhoneHomeID:    phoneHomeID,
		PhoneHomeURL: fmt.Sprintf(
			"%s/v1/installs/%s/phone-home/%s",
			a.cfg.PublicAPIURL,
			req.InstallID,
			phoneHomeID,
		),
		AWSBucketName: a.cfg.AWSCloudFormationStackTemplateBucket,
		AWSBucketKey:  bucketKey,
		TemplateURL:   templateURL,
		QuickLinkURL:  quickLinkURL,
	}

	if res := a.db.WithContext(ctx).Create(&obj); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create cloudformation stack version")
	}

	return &obj, nil
}
