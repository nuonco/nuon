package links

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

func InstallDeployLinks(cfg *internal.Config, installID, componentID, deployID string) map[string]any {
	return map[string]any{
		//"ui":  InstallDeployUILink(cfg, installID, componentID, deployID),
		//"api": InstallDeployAPILink(cfg, id),
	}
}

//func InstallDeployUILink(cfg *internal.Config, id string) string {
//link, err := url.JoinPath(cfg.AppURL,
//"installs",
//installDeploy.InstallID,
//"components",
//installDeploy.InstallComponentID,
//"deploys",
//installDeploy.ID,
//)
//if err != nil {
//return handleErr(cfg, err)
//}

//return link
//}

func InstallSignalLink(cfg *internal.Config, id string, typ string) string {
	return eventLoopSignalLink(cfg, "installs", id, typ)
}

func InstallEventLoopSignalLink(cfg *internal.Config, id string, typ string) string {
	return eventLoopLink(cfg, "installs", id)
}

//func InstallDeployAPILink(cfg *internal.Config, id string) string {
//link, err := url.JoinPath(cfg.PublicAPIURL,
//"v1",
//"installs",
//id.InstallID,
//"deploys",
//id.ID,
//)
//if err != nil {
//return handleErr(cfg, err)
//}

//return link
//}
