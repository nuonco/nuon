package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	githubclient "github.com/powertoolsdev/mono/services/api/internal/clients/github"
	temporalclient "github.com/powertoolsdev/mono/services/api/internal/clients/temporal"
	adminserver "github.com/powertoolsdev/mono/services/api/internal/servers/admin"
	appsserver "github.com/powertoolsdev/mono/services/api/internal/servers/apps"
	componentsserver "github.com/powertoolsdev/mono/services/api/internal/servers/components"
	deploymentsserver "github.com/powertoolsdev/mono/services/api/internal/servers/deployments"
	githubserver "github.com/powertoolsdev/mono/services/api/internal/servers/github"
	installsserver "github.com/powertoolsdev/mono/services/api/internal/servers/installs"
	orgsserver "github.com/powertoolsdev/mono/services/api/internal/servers/orgs"
	statusserver "github.com/powertoolsdev/mono/services/api/internal/servers/status"
	usersserver "github.com/powertoolsdev/mono/services/api/internal/servers/users"
	"github.com/powertoolsdev/mono/services/api/internal/services"
)

func (s *server) registerApps(mux *http.ServeMux) error {
	appsTc, err := temporalclient.New(temporalclient.WithConfig(s.cfg), temporalclient.WithNamespace("apps"))
	if err != nil {
		return fmt.Errorf("unable to create temporal client: %w", err)
	}
	appSvc := services.NewAppService(s.db, appsTc, s.log)
	_, err = appsserver.New(s.v,
		appsserver.WithHTTPMux(mux),
		appsserver.WithService(appSvc),
		appsserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize apps server: %w", err)
	}

	return nil
}

func (s *server) registerAdmin(mux *http.ServeMux) error {
	adminSvc := services.NewAdminService(s.db, s.log)
	_, err := adminserver.New(s.v,
		adminserver.WithHTTPMux(mux),
		adminserver.WithService(adminSvc),
		adminserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize admin server: %w", err)
	}

	return nil
}

func (s *server) registerComponents(mux *http.ServeMux) error {
	componentsSvc := services.NewComponentService(s.db, s.log)
	_, err := componentsserver.New(s.v,
		componentsserver.WithHTTPMux(mux),
		componentsserver.WithService(componentsSvc),
		componentsserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize components server: %w", err)
	}

	return nil
}

func (s *server) registerDeployments(mux *http.ServeMux) error {
	deploymentsTc, err := temporalclient.New(temporalclient.WithConfig(s.cfg), temporalclient.WithNamespace("deployments"))
	if err != nil {
		return fmt.Errorf("unable to create temporal client: %w", err)
	}

	ghTransport, err := githubclient.New(githubclient.WithConfig(s.cfg))
	if err != nil {
		return fmt.Errorf("unable to github client: %w", err)
	}

	deploymentsSvc := services.NewDeploymentService(s.db, deploymentsTc, ghTransport, s.cfg.GithubAppID, s.cfg.GithubAppKeySecretName, s.log)
	_, err = deploymentsserver.New(s.v,
		deploymentsserver.WithHTTPMux(mux),
		deploymentsserver.WithService(deploymentsSvc),
		deploymentsserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize deployments server: %w", err)
	}

	return nil
}

func (s *server) registerGithub(mux *http.ServeMux) error {
	// get github app details from config
	githubAppID, err := strconv.ParseInt(s.cfg.GithubAppID, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse github app id: %w", err)
	}

	appstp, err := ghinstallation.NewAppsTransport(http.DefaultTransport, githubAppID, []byte(s.cfg.GithubAppKey))
	if err != nil {
		return fmt.Errorf("unable to parse github app id: %w", err)
	}
	githubSvc := services.NewGithubService(appstp, s.log)

	_, err = githubserver.New(s.v,
		githubserver.WithHTTPMux(mux),
		githubserver.WithService(githubSvc),
		githubserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize github server: %w", err)
	}

	return nil
}

func (s *server) registerInstalls(mux *http.ServeMux) error {
	installsTc, err := temporalclient.New(temporalclient.WithConfig(s.cfg), temporalclient.WithNamespace("installs"))
	if err != nil {
		return fmt.Errorf("unable to create temporal client: %w", err)
	}
	installSvc := services.NewInstallService(s.db, installsTc, s.log)
	_, err = installsserver.New(s.v,
		installsserver.WithHTTPMux(mux),
		installsserver.WithService(installSvc),
		installsserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}
	return nil
}

func (s *server) registerOrgs(mux *http.ServeMux) error {
	orgsTc, err := temporalclient.New(temporalclient.WithConfig(s.cfg), temporalclient.WithNamespace("orgs"))
	if err != nil {
		return fmt.Errorf("unable to create temporal client: %w", err)
	}
	orgSvc := services.NewOrgService(s.db, orgsTc, s.log)
	_, err = orgsserver.New(s.v,
		orgsserver.WithHTTPMux(mux),
		orgsserver.WithService(orgSvc),
		orgsserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize orgs server: %w", err)
	}
	return nil
}

func (s *server) registerUsers(mux *http.ServeMux) error {
	userSvc := services.NewUserService(s.db, s.log)
	_, err := usersserver.New(s.v,
		usersserver.WithHTTPMux(mux),
		usersserver.WithService(userSvc),
		usersserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize users server: %w", err)
	}

	return nil
}

func (s *server) registerStatus(mux *http.ServeMux) error {
	_, err := statusserver.New(s.v,
		statusserver.WithHTTPMux(mux),
		statusserver.WithGitRef(s.cfg.GitRef),
		statusserver.WithInterceptors(s.interceptors...),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize status server: %w", err)
	}
	return nil
}

func (s *server) registerLoadbalancerHealthCheck(mux *http.ServeMux) {
	mux.Handle("/_ping", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write([]byte("{\"status\": \"ok\"}")); err != nil {
			log.Fatal("unable to write load balancer health check response", err.Error())
		}
	}))
}

func (s *server) registerReflectServer(mux *http.ServeMux) {
	reflector := grpcreflect.NewStaticReflector(srvs...)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
}

func (s *server) registerAll(mux *http.ServeMux) error {
	if err := s.registerApps(mux); err != nil {
		return fmt.Errorf("unable to register apps: %w", err)
	}

	if err := s.registerAdmin(mux); err != nil {
		return fmt.Errorf("unable to register admin: %w", err)
	}

	if err := s.registerComponents(mux); err != nil {
		return fmt.Errorf("unable to register components: %w", err)
	}

	if err := s.registerDeployments(mux); err != nil {
		return fmt.Errorf("unable to register deployments: %w", err)
	}

	if err := s.registerGithub(mux); err != nil {
		return fmt.Errorf("unable to register github: %w", err)
	}

	if err := s.registerInstalls(mux); err != nil {
		return fmt.Errorf("unable to register installs: %w", err)
	}

	if err := s.registerOrgs(mux); err != nil {
		return fmt.Errorf("unable to register orgs: %w", err)
	}

	if err := s.registerUsers(mux); err != nil {
		return fmt.Errorf("unable to register users: %w", err)
	}

	if err := s.registerStatus(mux); err != nil {
		return fmt.Errorf("unable to register status: %w", err)
	}

	s.registerLoadbalancerHealthCheck(mux)
	s.registerReflectServer(mux)
	return nil
}
