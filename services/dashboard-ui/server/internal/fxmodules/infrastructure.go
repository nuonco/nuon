package fxmodules

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/agent"
)

func newLogger(cfg *internal.Config) (*zap.Logger, error) {
	if cfg.LogLevel == "DEBUG" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func newAgentConfig(cfg *internal.Config) *agent.Config {
	return &agent.Config{
		Enabled:      cfg.AgentEnabled,
		ProviderName: cfg.AgentProvider,
		Model:        cfg.AgentModel,
		MaxTurns:     cfg.AgentMaxTurns,
		AnthropicKey: cfg.AnthropicAPIKey,
	}
}

func newAgentDeps(cfg *internal.Config, agentCfg *agent.Config, l *zap.Logger) (*agent.Agent, *agent.ConversationStore) {
	store := agent.NewConversationStore()

	if !agentCfg.Enabled {
		return agent.NewAgent(nil, nil, store, agentCfg, l), store
	}

	provider, err := agent.NewProvider(agentCfg)
	if err != nil {
		l.Warn("agent provider initialization failed, disabling agent", zap.Error(err))
		agentCfg.Enabled = false
		return agent.NewAgent(nil, nil, store, agentCfg, l), store
	}

	executor := agent.NewToolExecutor(cfg.APIUrl)
	return agent.NewAgent(provider, executor, store, agentCfg, l), store
}

var InfrastructureModule = fx.Module("infrastructure",
	fx.Provide(internal.NewConfig),
	fx.Provide(newLogger),
	fx.Provide(NewMetricsWriter),
	fx.Provide(newAgentConfig),
	fx.Provide(newAgentDeps),
)
