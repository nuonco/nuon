package app

type InstallStackVersionSecret struct {
	Name        string `json:"name"`
	Description string `json:"description,omitzero"`
	Required    bool   `json:"required"`
}

// CustomerStackSecrets expects secrets already rendered against install state.
func CustomerStackSecrets(secrets []AppSecretConfig) []InstallStackVersionSecret {
	var out []InstallStackVersionSecret
	for _, sec := range secrets {
		if sec.AutoGenerate {
			continue
		}
		out = append(out, InstallStackVersionSecret{
			Name:        sec.Name,
			Description: sec.Description,
			Required:    sec.Required,
		})
	}
	return out
}
