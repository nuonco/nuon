package oapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	accountsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/service"
	actionsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/service"
	appsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/service"
	awsaccountconnectionsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/aws-account-connections/service"
	componentsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/components/service"
	generalservice "github.com/nuonco/nuon/services/ctl-api/internal/app/general/service"
	identityprovidersservice "github.com/nuonco/nuon/services/ctl-api/internal/app/identity-providers/service"
	installsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/service"
	notebooksservice "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/service"
	oidcfederationservice "github.com/nuonco/nuon/services/ctl-api/internal/app/oidc-federation/service"
	onboardingservice "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/service"
	orgsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/service"
	policyreportsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/service"
	queuesservice "github.com/nuonco/nuon/services/ctl-api/internal/app/queues/service"
	runbooksservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/service"
	runnerauthservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runner-auth/service"
	runnersservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/service"
	slackservice "github.com/nuonco/nuon/services/ctl-api/internal/app/slack/service"
	stacksservice "github.com/nuonco/nuon/services/ctl-api/internal/app/stacks/service"
	vcsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/service"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"

	"github.com/nuonco/nuon/services/ctl-api/docs/admin"
	"github.com/nuonco/nuon/services/ctl-api/docs/public"
	"github.com/nuonco/nuon/services/ctl-api/docs/runner"
)

func testDomainServices(t *testing.T, ea *api.EndpointAudit) []api.Service {
	t.Helper()
	runnersService, err := runnersservice.New(runnersservice.Params{EndpointAudit: ea})
	require.NoError(t, err)

	return []api.Service{
		accountsservice.New(accountsservice.Params{}),
		actionsservice.New(actionsservice.Params{EndpointAudit: ea}),
		awsaccountconnectionsservice.New(awsaccountconnectionsservice.Params{EndpointAudit: ea}),
		appsservice.New(appsservice.Params{EndpointAudit: ea}),
		componentsservice.New(componentsservice.Params{EndpointAudit: ea}),
		generalservice.New(generalservice.Params{EndpointAudit: ea}),
		identityprovidersservice.New(identityprovidersservice.Params{}),
		installsservice.New(installsservice.Params{EndpointAudit: ea}),
		notebooksservice.New(notebooksservice.Params{EndpointAudit: ea}),
		orgsservice.New(orgsservice.Params{EndpointAudit: ea}),
		policyreportsservice.New(policyreportsservice.Params{EndpointAudit: ea}),
		queuesservice.New(queuesservice.Params{}),
		runnerauthservice.New(runnerauthservice.Params{}),
		runnersService,
		runbooksservice.New(runbooksservice.Params{EndpointAudit: ea}),
		slackservice.New(slackservice.Params{
			EndpointAudit: ea,
			Cfg:           &internal.Config{SlackSigningSecret: "test-signing-secret"},
		}),
		stacksservice.New(stacksservice.Params{EndpointAudit: ea}),
		vcsservice.New(vcsservice.Params{}),
		onboardingservice.New(onboardingservice.Params{EndpointAudit: ea}),
		oidcfederationservice.New(oidcfederationservice.Params{}),
	}
}

type SwaggerSpec struct {
	Paths map[string]PathItem `json:"paths"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

type Operation struct {
	OperationID string      `json:"operationId"`
	Summary     string      `json:"summary"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func TestSwaggerPaginationConsistency(t *testing.T) {
	specs := []struct {
		name     string
		specJSON string
	}{
		{
			name:     "public",
			specJSON: public.SwaggerInfo.ReadDoc(),
		},
		{
			name:     "admin",
			specJSON: admin.SwaggerInfoadmin.ReadDoc(),
		},
		{
			name:     "runner",
			specJSON: runner.SwaggerInforunner.ReadDoc(),
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			swaggerSpec := parseSwaggerSpec(t, spec.specJSON, spec.name)

			var violations []string

			for path, pathItem := range swaggerSpec.Paths {
				if pathItem.Get != nil {
					violation := checkPaginationParams(path, pathItem.Get)
					if violation != "" {
						violations = append(violations, violation)
					}
				}
			}

			if len(violations) > 0 {
				errorMsg := fmt.Sprintf("Found %d pagination consistency violation(s) in %s spec:\n", len(violations), spec.name)
				for _, v := range violations {
					errorMsg += fmt.Sprintf("  - %s\n", v)
				}
				assert.Fail(t, errorMsg)
			}
		})
	}
}

func checkPaginationParams(path string, operation *Operation) string {
	paramMap := make(map[string]bool)
	allParamMap := make(map[string]bool)

	for _, param := range operation.Parameters {
		allParamMap[param.Name] = true
		if param.In == "query" {
			paramMap[param.Name] = true
		}
	}

	if paramMap["page"] {
		missingParams := []string{}

		if !paramMap["offset"] {
			missingParams = append(missingParams, "offset")
		}
		if !paramMap["limit"] {
			missingParams = append(missingParams, "limit")
		}

		if len(missingParams) > 0 {
			return fmt.Sprintf("GET %s (operation: %s) has 'page' parameter but missing: %v",
				path, operation.OperationID, missingParams)
		}

		if allParamMap["x-nuon-pagination-enabled"] {
			return fmt.Sprintf("GET %s (operation: %s) uses standard pagination (page/offset/limit) but also has 'x-nuon-pagination-enabled' parameter, which should not be used together",
				path, operation.OperationID)
		}
	}

	return ""
}

func parseSwaggerSpec(t *testing.T, specJSON string, specName string) *SwaggerSpec {
	var spec SwaggerSpec
	err := json.Unmarshal([]byte(specJSON), &spec)
	require.NoError(t, err, "Failed to parse swagger spec JSON for %s", specName)

	return &spec
}

func TestSwaggerRoutesRegisteredInGin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ea := api.NewEndpointAudit()
	services := testDomainServices(t, ea)

	specs := []struct {
		name      string
		specJSON  string
		registers []func(api.Service, *gin.Engine) error
	}{
		{
			name:     "public",
			specJSON: public.SwaggerInfo.ReadDoc(),
			registers: []func(api.Service, *gin.Engine) error{
				func(svc api.Service, e *gin.Engine) error { return svc.RegisterPublicRoutes(e) },
				func(svc api.Service, e *gin.Engine) error { return svc.RegisterAuthRoutes(e) },
				func(svc api.Service, e *gin.Engine) error { return svc.RegisterSlackRoutes(e) },
			},
		},
		{
			name:     "admin",
			specJSON: admin.SwaggerInfoadmin.ReadDoc(),
			registers: []func(api.Service, *gin.Engine) error{
				func(svc api.Service, e *gin.Engine) error { return svc.RegisterInternalRoutes(e) },
			},
		},
		{
			name:     "runner",
			specJSON: runner.SwaggerInforunner.ReadDoc(),
			registers: []func(api.Service, *gin.Engine) error{
				func(svc api.Service, e *gin.Engine) error { return svc.RegisterRunnerRoutes(e) },
			},
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			engine := gin.New()
			for _, svc := range services {
				for _, register := range spec.registers {
					err := register(svc, engine)
					require.NoError(t, err, "failed to register routes for %s", spec.name)
				}
			}

			ginRoutes := make(map[string]struct{})
			for _, route := range engine.Routes() {
				key := route.Method + " " + route.Path
				ginRoutes[key] = struct{}{}
			}

			swaggerSpec := parseSwaggerSpec(t, spec.specJSON, spec.name)

			var missing []string
			for path, pathItem := range swaggerSpec.Paths {
				ginPath := swaggerPathToGinPath(path)
				for _, method := range getPathMethods(pathItem) {
					key := method + " " + ginPath
					if _, ok := ginRoutes[key]; !ok {
						missing = append(missing, key)
					}
				}
			}

			if len(missing) > 0 {
				sort.Strings(missing)
				msg := fmt.Sprintf("Found %d swagger route(s) in %s spec not registered in gin:\n", len(missing), spec.name)
				for _, r := range missing {
					msg += fmt.Sprintf("  - %s\n", r)
				}
				assert.Fail(t, msg)
			}
		})
	}
}

func swaggerPathToGinPath(path string) string {
	var result strings.Builder
	for i := 0; i < len(path); i++ {
		if path[i] == '{' {
			result.WriteByte(':')
			i++
			for i < len(path) && path[i] != '}' {
				result.WriteByte(path[i])
				i++
			}
		} else {
			result.WriteByte(path[i])
		}
	}
	return result.String()
}

func getPathMethods(p PathItem) []string {
	var methods []string
	if p.Get != nil {
		methods = append(methods, "GET")
	}
	if p.Post != nil {
		methods = append(methods, "POST")
	}
	if p.Put != nil {
		methods = append(methods, "PUT")
	}
	if p.Patch != nil {
		methods = append(methods, "PATCH")
	}
	if p.Delete != nil {
		methods = append(methods, "DELETE")
	}
	return methods
}

func extractPathParams(path string) []string {
	var params []string
	for i := 0; i < len(path); i++ {
		if path[i] == '{' {
			i++
			start := i
			for i < len(path) && path[i] != '}' {
				i++
			}
			params = append(params, path[start:i])
		}
	}
	return params
}

func getOperations(p PathItem) map[string]*Operation {
	ops := make(map[string]*Operation)
	if p.Get != nil {
		ops["GET"] = p.Get
	}
	if p.Post != nil {
		ops["POST"] = p.Post
	}
	if p.Put != nil {
		ops["PUT"] = p.Put
	}
	if p.Patch != nil {
		ops["PATCH"] = p.Patch
	}
	if p.Delete != nil {
		ops["DELETE"] = p.Delete
	}
	return ops
}

func TestSwaggerParamNamesConsistency(t *testing.T) {
	specs := []struct {
		name     string
		specJSON string
	}{
		{
			name:     "public",
			specJSON: public.SwaggerInfo.ReadDoc(),
		},
		{
			name:     "admin",
			specJSON: admin.SwaggerInfoadmin.ReadDoc(),
		},
		{
			name:     "runner",
			specJSON: runner.SwaggerInforunner.ReadDoc(),
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			swaggerSpec := parseSwaggerSpec(t, spec.specJSON, spec.name)

			var violations []string

			for path, pathItem := range swaggerSpec.Paths {
				pathParams := extractPathParams(path)

				for method, op := range getOperations(pathItem) {
					annotatedParams := make(map[string]bool)
					for _, param := range op.Parameters {
						if param.In == "path" {
							annotatedParams[param.Name] = true
						}
					}

					urlParams := make(map[string]bool)
					for _, p := range pathParams {
						urlParams[p] = true
					}

					for _, p := range pathParams {
						if !annotatedParams[p] {
							violations = append(violations, fmt.Sprintf(
								"%s %s (operation: %s): path has {%s} but no @Param annotation with in=path for %q",
								method, path, op.OperationID, p, p))
						}
					}

					for name := range annotatedParams {
						if !urlParams[name] {
							violations = append(violations, fmt.Sprintf(
								"%s %s (operation: %s): @Param %q has in=path but {%s} not found in URL",
								method, path, op.OperationID, name, name))
						}
					}
				}
			}

			if len(violations) > 0 {
				sort.Strings(violations)
				msg := fmt.Sprintf("Found %d @Param/path inconsistency(ies) in %s spec:\n", len(violations), spec.name)
				for _, v := range violations {
					msg += fmt.Sprintf("  - %s\n", v)
				}
				assert.Fail(t, msg)
			}
		})
	}
}

func TestSwaggerSpecsExist(t *testing.T) {
	specs := []string{
		"../../../docs/public/swagger.json",
		"../../../docs/public/docs.go",
		"../../../docs/admin/admin_swagger.json",
		"../../../docs/admin/admin_docs.go",
		"../../../docs/runner/runner_swagger.json",
		"../../../docs/runner/runner_docs.go",
	}

	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			absPath, err := filepath.Abs(spec)
			require.NoError(t, err)

			info, err := os.Stat(absPath)
			require.NoError(t, err, "Swagger spec file does not exist: %s. Run 'go generate' to create it.", absPath)
			require.False(t, info.IsDir(), "Expected file but found directory: %s", absPath)
		})
	}
}
