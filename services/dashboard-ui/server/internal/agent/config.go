package agent

import "fmt"

type Config struct {
	Enabled      bool   `config:"agent_enabled"`
	ProviderName string `config:"agent_provider"`
	Model        string `config:"agent_model"`
	MaxTurns     int    `config:"agent_max_turns"`
	AnthropicKey string `config:"anthropic_api_key"`
}

func (c *Config) Defaults() {
	if c.ProviderName == "" {
		c.ProviderName = "anthropic"
	}
	if c.MaxTurns == 0 {
		c.MaxTurns = 10
	}
}

func NewProvider(cfg *Config) (Provider, error) {
	cfg.Defaults()

	if !cfg.Enabled {
		return nil, nil
	}

	switch cfg.ProviderName {
	case "anthropic":
		if cfg.AnthropicKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when agent_provider=anthropic")
		}
		return NewAnthropicProvider(cfg.AnthropicKey, cfg.Model), nil
	default:
		return nil, fmt.Errorf("unknown agent provider: %s", cfg.ProviderName)
	}
}
