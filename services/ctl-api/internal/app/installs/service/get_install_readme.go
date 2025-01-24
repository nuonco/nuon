package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type Readme struct {
	ReadMe string `json:"readme"`
}

// @ID GetInstallReadme
// @Summary	get install readme rendered with
// @Description.markdown	get_install_readme.md
// @Param			install_id	path	string	true	"install ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object} Readme
// @Router			/v1/installs/{install_id}/readme [get]
func (s *service) GetInstallReadme(ctx *gin.Context) {
	// get install state
	installID := ctx.Param("install_id")
	installState, err := s.getInstallState(ctx, installID)
	if err != nil {
		response := Readme{err.Error()}
		ctx.JSON(http.StatusOK, response)
		ctx.Error(fmt.Errorf("unable to get install state: %w", err))
		return
	}

	// get app readme template
	appConfig, err := s.getLatestAppConfig(ctx, installState.App.ID)
	if err != nil {
		response := Readme{err.Error()}
		ctx.JSON(http.StatusInternalServerError, response)
		ctx.Error(fmt.Errorf("unable to get latest app config: %w", err))
		return
	}

	// interpolate the state into the readme md
	value, err := renderString(appConfig.Readme, *installState)
	if err != nil {
		// TODO(ja):
		// If we can't interpolate the README contents, we still want to return the un-rendered README template,
		// so clients can provide it as context for what failed to render.
		response := Readme{value}
		ctx.JSON(http.StatusOK, response)
		// TODO(ja):
		// If we set this, it causes the HTTP response to have an error status, which causes the JS async call to throw,
		// which prevents the client from getting the README template. We should re-think how we handle API error responses
		// to get around this.
		// ctx.Error(fmt.Errorf("unable to render readme: %w", err))
		return
	}

	response := Readme{value}
	ctx.JSON(http.StatusOK, response)
}

type Data struct {
	Nuon state.InstallState `json:"nuon"`
}

func renderString(inputVal string, installState state.InstallState) (string, error) {
	// if the README template is empty, return empty string
	if inputVal == "" {
		return "", nil
	}

	// format install state data to be used as variables
	// need to format as JSON to lowercase all the fields
	data := Data{
		Nuon: installState,
	}
	jsonString, err := json.Marshal(data)
	if err != nil {
		return inputVal, err
	}
	parsedJSON := make(map[string]interface{})
	err = json.Unmarshal(jsonString, &parsedJSON)
	if err != nil {
		return inputVal, err
	}

	// render the template
	temp, err := template.New("input").Option("missingkey=zero").Parse(inputVal)
	if err != nil {
		return inputVal, nil
	}
	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, parsedJSON); err != nil {
		return inputVal, fmt.Errorf("unable to execute template: %w", err)
	}
	outputVal := buf.String()
	if outputVal == "" {
		return inputVal, fmt.Errorf("rendered value was empty, this usually means a bad interpolation config: %s", inputVal)
	}
	if outputVal == "<no value>" {
		return inputVal, fmt.Errorf("rendered value was empty, which usually means a bad interpolation config: %s", inputVal)
	}

	return outputVal, nil
}

func (s *service) getLatestAppConfig(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig
	res := s.db.WithContext(ctx).Where("app_id = ?", appID).Order("created_at DESC").First(&appConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}
	return &appConfig, nil
}
