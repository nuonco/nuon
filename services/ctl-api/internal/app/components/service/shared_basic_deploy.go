package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type basicDeployConfigRequest struct {
	InstanceCount   int                `json:"instance_count"`
	ListenPort      int                `json:"listen_port"`
	HealthCheckPath string             `json:"health_check_path"`
	CPURequest      string             `json:"cpu_request"`
	CPULimit        string             `json:"cpu_limit"`
	MemRequest      string             `json:"mem_request"`
	MemLimit        string             `json:"mem_limit"`
	EnvVars         map[string]*string `json:"env_vars"`
	Args            []string           `json:"args"`
}

func (b *basicDeployConfigRequest) getBasicDeployConfig() *app.BasicDeployConfig {
	if b == nil {
		return nil
	}

	return &app.BasicDeployConfig{
		InstanceCount:   b.InstanceCount,
		ListenPort:      b.ListenPort,
		HealthCheckPath: b.HealthCheckPath,
		CPURequest:      b.CPURequest,
		CPULimit:        b.CPULimit,
		MemRequest:      b.MemRequest,
		MemLimit:        b.MemLimit,
		EnvVars:         pgtype.Hstore(b.EnvVars),
		Args:            b.Args,
	}
}
