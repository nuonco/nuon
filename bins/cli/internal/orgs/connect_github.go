package orgs

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/browser"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ConnectGithub(ctx context.Context) error {
	var v *models.AppVCSConnection
	var connection = "started"
	view := ui.NewGetView()

	ogVcs, err := s.api.GetVCSConnections(ctx)
	if err != nil {
		view.Error(err)
		connection = "failed"
	}

	timeout := 1
	var connected bool
	for !connected {
		fmt.Printf("connection: %s\n", connection)

		if connection == "started" {
			connection = "pending"
			browser.OpenURL("https://github.com/apps/" + s.cfg.GitHubAppName + "/installations/new?state=" + s.cfg.OrgID)
		}

		vcs, err := s.api.GetVCSConnections(ctx)
		if err != nil {
			view.Error(err)
			connection = "failed"
		}

		if len(vcs) > len(ogVcs) || connection == "failed" {
			v = vcs[len(vcs)-1]
			fmt.Println("connection: connected")
			connection = "connected"
			connected = true
		} else {
			time.Sleep(5 * time.Second)
		}

		if timeout == 18 {
			fmt.Println("connection: time out")
			connection = "time out"
			connected = true
		}

		timeout++
	}

	if connection == "connected" {
		fmt.Println("")
		fmt.Println("new vcs connection")
		fmt.Println("-------------------------------------------------")
		view.Render([][]string{
			{"id", v.ID},
			{"github install id", v.GithubInstallID},
		})
	} else {
		fmt.Println("")
		fmt.Println("please try again")
	}
	return nil
}
