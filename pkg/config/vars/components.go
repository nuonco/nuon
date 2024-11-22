package vars

func (v *varsValidator) getComponents() map[string]*instanceIntermediate {
	comps := make(map[string]*instanceIntermediate, 0)

	for _, c := range v.cfg.Components {
		if c.Name == v.ignoreComponent {
			continue
		}

		name := c.Name
		if c.VarName != "" {
			name = c.VarName
		}

		comps[name] = &instanceIntermediate{}
	}

	return nil
}
