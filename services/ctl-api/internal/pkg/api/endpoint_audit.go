package api

type EndpointAudit struct {
	Routes map[string]struct{}
}

func (d *EndpointAudit) IsDeprecated(method, name, path string) bool {
	key := method + "_" + name + "_" + path

	if _, exists := d.Routes[key]; exists {
		return true
	}

	return false
}

func (d *EndpointAudit) Add(method, name, path string) {
	key := method + "_" + name + "_" + path
	d.Routes[key] = struct{}{}
}

var defaultEndpointAudit *EndpointAudit

func NewEndpointAudit() *EndpointAudit {
	if defaultEndpointAudit != nil {
		return defaultEndpointAudit
	}

	routes := make(map[string]struct{})
	defaultEndpointAudit = &EndpointAudit{
		Routes: routes,
	}

	return defaultEndpointAudit
}
