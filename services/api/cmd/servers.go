package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	githubclient "github.com/powertoolsdev/mono/services/api/internal/clients/github"
	adminserver "github.com/powertoolsdev/mono/services/api/internal/servers/admin"
	appsserver "github.com/powertoolsdev/mono/services/api/internal/servers/apps"
	buildsserver "github.com/powertoolsdev/mono/services/api/internal/servers/builds"
	componentsserver "github.com/powertoolsdev/mono/services/api/internal/servers/components"
	deployserver "github.com/powertoolsdev/mono/services/api/internal/servers/deploy"
	deploymentsserver "github.com/powertoolsdev/mono/services/api/internal/servers/deployments"
	githubserver "github.com/powertoolsdev/mono/services/api/internal/servers/github"
	installsserver "github.com/powertoolsdev/mono/services/api/internal/servers/installs"
	instancesserver "github.com/powertoolsdev/mono/services/api/internal/servers/instances"
	orgsserver "github.com/powertoolsdev/mono/services/api/internal/servers/orgs"
	statusserver "github.com/powertoolsdev/mono/services/api/internal/servers/status"
	usersserver "github.com/powertoolsdev/mono/services/api/internal/servers/users"
	"github.com/powertoolsdev/mono/services/api/internal/services"
)

func (a *app) registerAppsServer(mux *http.ServeMux) error {
	appSvc := services.NewAppService(a.db, a.tc, a.log)
	_, err := appsserver.New(a.v,
		appsserver.WithService(appSvc),
		appsserver.WithInterceptors(a.interceptors...),
		appsserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize apps server: %w", err)
	}

	return nil
}

func (a *app) registerAdminServer(mux *http.ServeMux) error {
	adminSvc := services.NewAdminService(a.db, a.log)
	_, err := adminserver.New(a.v,
		adminserver.WithService(adminSvc),
		adminserver.WithInterceptors(a.interceptors...),
		adminserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize admin server: %w", err)
	}

	return nil
}

func (a *app) registerComponentsServer(mux *http.ServeMux) error {
	componentsSvc := services.NewComponentService(a.db, a.log)
	_, err := componentsserver.New(a.v,
		componentsserver.WithService(componentsSvc),
		componentsserver.WithInterceptors(a.interceptors...),
		componentsserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize components server: %w", err)
	}

	return nil
}

func (a *app) registerDeploymentsServer(mux *http.ServeMux) error {
	ghTransport, err := githubclient.New(githubclient.WithConfig(a.cfg))
	if err != nil {
		return fmt.Errorf("unable to github client: %w", err)
	}

	deploymentsSvc := services.NewDeploymentService(a.db, a.tc, ghTransport, a.cfg.GithubAppID, a.cfg.GithubAppKeySecretName, a.log)
	_, err = deploymentsserver.New(a.v,
		deploymentsserver.WithService(deploymentsSvc),
		deploymentsserver.WithInterceptors(a.interceptors...),
		deploymentsserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize deployments server: %w", err)
	}

	return nil
}

func (a *app) registerBuildsServer(mux *http.ServeMux) error {
	_, err := buildsserver.New(a.v,
		buildsserver.WithTemporalClient(a.tc),
		buildsserver.WithGithubClient(a.cfg),
		buildsserver.WithInterceptors(a.interceptors...),
		buildsserver.WithHTTPMux(mux),
		buildsserver.WithDBClient(a.db),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize builds server: %w", err)
	}

	return nil
}
func (a *app) registerDeployServer(mux *http.ServeMux) error {
	_, err := deployserver.New(a.v,
		deployserver.WithInterceptors(a.interceptors...),
		deployserver.WithHTTPMux(mux),
		deployserver.WithDB(a.db),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize deploy server: %w", err)
	}

	return nil
}

func (a *app) registerGithubServer(mux *http.ServeMux) error {
	// get github app details from config
	githubAppID, err := strconv.ParseInt(a.cfg.GithubAppID, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse github app id: %w", err)
	}

	appstp, err := ghinstallation.NewAppsTransport(http.DefaultTransport, githubAppID, []byte(a.cfg.GithubAppKey))
	if err != nil {
		return fmt.Errorf("unable to parse github app id: %w", err)
	}
	githubSvc := services.NewGithubService(appstp, a.log)

	_, err = githubserver.New(a.v,
		githubserver.WithService(githubSvc),
		githubserver.WithInterceptors(a.interceptors...),
		githubserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize github server: %w", err)
	}

	return nil
}

func (a *app) registerInstallsServer(mux *http.ServeMux) error {
	installSvc := services.NewInstallService(a.db, a.tc, a.log)
	_, err := installsserver.New(a.v,
		installsserver.WithService(installSvc),
		installsserver.WithInterceptors(a.interceptors...),
		installsserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}
	return nil
}

func (a *app) registerInstancesServer(mux *http.ServeMux) error {
	instanceSvc := services.NewInstanceService(a.db, a.log)
	_, err := instancesserver.New(a.v,
		instancesserver.WithService(instanceSvc),
		instancesserver.WithInterceptors(a.interceptors...),
		instancesserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize instances server: %w", err)
	}
	return nil
}

func (a *app) registerOrgsServer(mux *http.ServeMux) error {
	orgSvc := services.NewOrgService(a.db, a.tc, a.log)
	_, err := orgsserver.New(a.v,
		orgsserver.WithService(orgSvc),
		orgsserver.WithInterceptors(a.interceptors...),
		orgsserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize orgs server: %w", err)
	}
	return nil
}

func (a *app) registerUsersServer(mux *http.ServeMux) error {
	userSvc := services.NewUserService(a.db, a.log)
	_, err := usersserver.New(a.v,
		usersserver.WithService(userSvc),
		usersserver.WithInterceptors(a.interceptors...),
		usersserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize users server: %w", err)
	}

	return nil
}

func (a *app) registerStatusServer(mux *http.ServeMux) error {
	_, err := statusserver.New(a.v,
		statusserver.WithGitRef(a.cfg.GitRef),
		statusserver.WithInterceptors(a.interceptors...),
		statusserver.WithHTTPMux(mux),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize status server: %w", err)
	}
	return nil
}

func (a *app) registerLoadbalancerHealthCheck(mux *http.ServeMux) {
	mux.Handle("/_ping", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write([]byte("{\"status\": \"ok\"}")); err != nil {
			log.Fatal("unable to write load balancer health check response", err.Error())
		}
	}))
}

func (a *app) registerReflectServer(mux *http.ServeMux) {
	reflector := grpcreflect.NewStaticReflector(srvs...)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
}

func (a *app) registerAllServers(mux *http.ServeMux) error {
	if err := a.registerAppsServer(mux); err != nil {
		return fmt.Errorf("unable to register apps: %w", err)
	}

	if err := a.registerAdminServer(mux); err != nil {
		return fmt.Errorf("unable to register admin: %w", err)
	}

	if err := a.registerComponentsServer(mux); err != nil {
		return fmt.Errorf("unable to register components: %w", err)
	}

	if err := a.registerDeploymentsServer(mux); err != nil {
		return fmt.Errorf("unable to register deployments: %w", err)
	}

	if err := a.registerBuildsServer(mux); err != nil {
		return fmt.Errorf("unable to register deployments: %w", err)
	}

	if err := a.registerDeployServer(mux); err != nil {
		return fmt.Errorf("unable to register deploys: %w", err)
	}

	if err := a.registerGithubServer(mux); err != nil {
		return fmt.Errorf("unable to register github: %w", err)
	}

	if err := a.registerInstallsServer(mux); err != nil {
		return fmt.Errorf("unable to register installs: %w", err)
	}

	if err := a.registerInstancesServer(mux); err != nil {
		return fmt.Errorf("unable to register instances: %w", err)
	}

	if err := a.registerOrgsServer(mux); err != nil {
		return fmt.Errorf("unable to register orgs: %w", err)
	}

	if err := a.registerUsersServer(mux); err != nil {
		return fmt.Errorf("unable to register users: %w", err)
	}

	if err := a.registerStatusServer(mux); err != nil {
		return fmt.Errorf("unable to register status: %w", err)
	}

	a.registerLoadbalancerHealthCheck(mux)
	a.registerReflectServer(mux)
	return nil
}
