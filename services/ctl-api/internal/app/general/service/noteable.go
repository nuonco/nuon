package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *service) MigrateNoteable(ctx *gin.Context) {
	ctx.Set("user_id", "google-oauth2|103491535260141311695")
	ctx.Set("org_id", "orgtm2hctoi1rjmfpwslddwrs7")

	sandbox := app.Sandbox{}
	res := s.db.WithContext(ctx).
		Preload("Releases").
		Where(app.Sandbox{
			Name: "aws-eks",
		}).First(&sandbox)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get sandbox release: %w", res.Error))
		return
	}

	org := app.Org{
		ID:   "orgtm2hctoi1rjmfpwslddwrs7",
		Name: "Noteable",
		VCSConnections: []app.VCSConnection{
			{
				GithubInstallID: "40074668",
			},
		},
		Status:            "active",
		StatusDescription: "active",
		Apps: []app.App{
			{
				ID:                "appjfasx3dh0nmpasct1y7w5ij",
				Name:              "Example App",
				Status:            "active",
				StatusDescription: "active",
				SandboxReleaseID:  sandbox.Releases[0].ID,
				Installs: []app.Install{
					{
						ID:                "inlcnurwp46ngprak2h4hbd4b6",
						Name:              "Customer One",
						SandboxReleaseID:  sandbox.Releases[0].ID,
						Status:            "active",
						StatusDescription: "active",
						AWSAccount: app.AWSAccount{
							IAMRoleARN: "arn:aws:iam::949309607565:role/nuon-demo-install-access",
							Region:     "us-west-2",
						},
					},
					{
						ID:                "inlutwbn9k9waesl2ur166rfen",
						Name:              "Customer Two",
						SandboxReleaseID:  sandbox.Releases[0].ID,
						Status:            "active",
						StatusDescription: "active",
						AWSAccount: app.AWSAccount{
							IAMRoleARN: "arn:aws:iam::949309607565:role/nuon-demo-install-access",
							Region:     "us-east-1",
						},
					},
				},
			},
			{
				ID:                "appomkchol8x5y4nv6smkwjoke",
				Name:              "Noteable Platform",
				Status:            "active",
				StatusDescription: "active",
				SandboxReleaseID:  sandbox.Releases[0].ID,
				Installs: []app.Install{
					{
						ID:                "inlq2djdgb44b522qxg02ktqhj",
						Name:              "AWS Integration Customer Simulation",
						SandboxReleaseID:  sandbox.Releases[0].ID,
						Status:            "active",
						StatusDescription: "active",
						AWSAccount: app.AWSAccount{
							IAMRoleARN: "arn:aws:iam::526968036384:role/integration-yeet-nuon-gg78k7vp",
							Region:     "us-east-2",
						},
					},
				},
			},
		},
	}

	res = s.db.WithContext(ctx).Create(&org)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to upsert data: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}
