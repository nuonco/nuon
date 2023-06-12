package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var runServerCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Run:   runServer,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(runServerCmd)
}

var srvs []string = []string{
	// shared handlers
	"shared.v1.StatusService",

	// local handlers
	"admin.v1.AdminService",
	"app.v1.AppService",
	"component.v1.ComponentService",
	"deployment.v1.DeploymentService",
	"github.v1.GithubService",
	"install.v1.InstallService",
	"instance.v1.InstanceService",
	"org.v1.OrgService",
	"user.v1.UserService",
}

//nolint:all
func runServer(cmd *cobra.Command, _ []string) {
	app, err := newApp(cmd.Flags())
	if err != nil {
		log.Fatalf("unable to load server: %s", err)
	}

	mux := http.NewServeMux()
	if err := app.registerAllServers(mux); err != nil {
		log.Fatalf("unable to register servers: %s", err)
	}

	app.log.Info("server starting: ",
		zap.String("host", app.cfg.HTTPAddress),
		zap.String("port", app.cfg.HTTPPort))

	if err := http.ListenAndServe(
		fmt.Sprintf("%s:%s", app.cfg.HTTPAddress, app.cfg.HTTPPort),

		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		app.log.Fatal("error on listen and server", zap.Error(err))
	}
}
