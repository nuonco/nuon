package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration035Installers(ctx context.Context) error {
	var appInstallers []*app.AppInstaller
	res := a.db.WithContext(ctx).
		Preload("App").
		Preload("Metadata").
		Order("created_at desc").
		Find(&appInstallers)
	if res.Error != nil {
		return res.Error
	}

	orgInstallers := make(map[string][]*app.AppInstaller, 0)
	for _, installer := range appInstallers {
		if _, ok := orgInstallers[installer.OrgID]; !ok {
			orgInstallers[installer.OrgID] = make([]*app.AppInstaller, 0)
		}

		orgInstallers[installer.OrgID] = append(orgInstallers[installer.OrgID], installer)
	}

	for orgID, appInstallers := range orgInstallers {
		inst := appInstallers[0]

		installer := &app.Installer{
			CreatedByID: inst.CreatedByID,
			OrgID:       orgID,
			Apps:        make([]app.App, 0),
			Type:        app.InstallerTypeSelfHosted,
			Metadata: app.InstallerMetadata{
				CreatedByID:         inst.CreatedByID,
				OrgID:               inst.OrgID,
				Name:                inst.Metadata.Name,
				Description:         inst.Metadata.Description,
				PostInstallMarkdown: inst.Metadata.PostInstallMarkdown,
				CopyrightMarkdown:   "",
				FooterMarkdown:      "",
				DocumentationURL:    inst.Metadata.DocumentationURL,
				LogoURL:             inst.Metadata.LogoURL,
				GithubURL:           inst.Metadata.GithubURL,
				CommunityURL:        inst.Metadata.CommunityURL,
				HomepageURL:         inst.Metadata.HomepageURL,
				DemoURL:             inst.Metadata.DemoURL,
				FaviconURL:          "data:image/x-icon;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAACXBIWXMAAAsTAAALEwEAmpwYAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAVwSURBVHgBrZd9TFVlHMe/5zn38hKoB0vS1DxMalMLLrWlhcllqaH9AWZkzgyxVTaXUn85zYCcL6sZGBa1poKmLNYU56bkLGGAhqsAh86BeC6k3oua9537es6vA4y5O+4byme7d+c+5/d7ft/zfd7u4RAl+kNmIU5BHovRZkFDusQEiAnxnACeMECwOGTFAJ5rhwYNWkU+WZedZImmXy5SQM5us0i8tihxElew5DleSJvBI20qj4SYwFSnn3DDpuCc0YtOi4x+N1VpeblUFWJ4KAH6IrMQEx9b/GQyV7RhuQavPKvBWDh304djPR7ccVO50yWXNqyIzpEhcta5xNc+dknV9T56VI5e89DyOpuUc8IsRlVcn+/Q5X7oltqvyTRe9Jhlevu4Q1pS69CFL653ibnveCVjv0LjjckuU36NU8qpDOGEXkdC1steqbv74Yq3G2T64aKbTI7Q+R23/JRT6ZT0ZWZhpC4buVA8SvG693gxNTXiwgjA1A9s3+vHkVN+JMfy+KbZhbob3qCxaU/xWJOuFWOV2OKAGwsEEvMzxz7m/1xSaP27fmq7/OCpHR6Ffu50U2GjPWTe2v0uyikZHoohB3hFKSncyDBWTH3AwvkMuucfuDa4P6yZF4spHj5k3kfZWshebdGQAD1ImDaFFSxbjYi0NQBb3gROH6bhBj9Awd0G/Rd6KDPn8kiUNQX6IhJYHO/Oe2mRH+FwWAj7PyUcLAXeKFCFNHLY8hbB2KPe9AXPUe6Fd3SxjheY053H4phbv6wg/MTb+KIHz+gIFeeBV3OBbQeAVZ9wOF017MrgUIwSYA3fZ3qqKtDOslgC509PSQ+vNo7z4/WCwJiMLOD9zzmQat7ODYQzNRRwn+wRBMxhIAenY0887hUThPDBcSy4z6TWzFjIYWslh+4OID9DQVM9RSVAPU2RyCCyyYJfQARkmxt9HaPnifr8Q6fZtFnApj0cvj3J0HQG2PWZArsp8n4SDyaweOaNGJjIPKjbacePH9hxx6CMEjHCtKeBrfs4ZMzncLuLIvaLAQ6MOdwRj8knJnuxqTYJ8xZpsGOpFb/scA+1M6IAASMsW8UhdXZkBxxGWBizugwea/hlOJEfLrhwbTwqupLU3YuwbfEAWk/51GsleJI3vACHFbDeZIbBZdhhbLGGF8BcAb/zt8dj84E48GrxPw770W8IdKF2n7o0r4ftEt1t6tcA2lki72u4e+F+2OBJ/MCotuRZDJm5DC/oCaUrvfh6vQ/NdYTCDILJABy8FN6B31XhWp/cwGQfq7tV2xd2HkzWOIO285yC6SkKvv87FulZDGcPy9h9gsOmMmBChLXV1TQ4rPaTrNCSbWE2Z7Wp3hRaAB9cgIaTEaN+BllawOPL4xpMFRGRpiNuOPs8VQ1IsgxtbzGcUt5T3BZaQJIP98v/Cmi7UnMPzXtMENN4jJX6XWbEMF9poKoZx8p7914OeYZbylqpK7OGbv10jRrzLlLL1uvksvhprJzdeY+KEq6Xj1LVJh4SLs2slmwXjCGTvf/a6GZJK91vuUMPg9TkoC8mdkrqZfAZoooQr6YdlQYLjTfWXjdVpLRJu4U2EeHoEr/T9aRXSe4rd2m8MLaYqWr2BalMaNYhGiSxUuxNqZAs5a30qFz5qouOJZ+VDgnnxWC1uNAiygQtQ4l2+qTNEzYvQPzKORgLd3+7javbO2Htc+9T/3GXrLCssIxJwAhGsUxknFyimSkUaObPRNzKudDOnQJuQmxAnGzzwPlnP6wX+9H/a69lwOyrVmSuPNu02hCu/6hfAsxiiSDzj+UpYHoiPh0TY0WaEC/I6nHks/ktXqvf4APfLstco8yzugxDYVQvo/8DuDhCmK3mGloAAAAASUVORK5CYII=",
			},
		}
		for _, appInstaller := range appInstallers {
			installer.Apps = append(installer.Apps, appInstaller.App)
		}

		// create installer
		res = a.db.WithContext(ctx).Create(&installer)
		if res.Error != nil {
			return fmt.Errorf("unable to create installer: %w", res.Error)
		}
	}

	return nil
}
