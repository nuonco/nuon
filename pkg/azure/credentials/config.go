package credentials

type ServicePrincipalCredentials struct {
	SubscriptionID       string `cty:"subscription_id" json:"subscription_id" temporaljson:"subscription_id" hcl:"subscription_id" mapstructure:"subscription_id,omitempty"`
	SubscriptionTenantID string `cty:"subscription_tenant_id" json:"subscription_tenant_id" temporaljson:"subscription_tenant_id" hcl:"subscription_tenant_id" mapstructure:"subscription_tenant_id,omitempty"`
}

type Config struct {
	ServicePrincipal *ServicePrincipalCredentials `cty:"service_principal,block" hcl:"service_principal,block" mapstructure:"service_principal,omitempty" json:"service_principal" temporaljson:"service_principal"`
	UseDefault       bool                         `cty:"use_default,optional" hcl:"use_default,optional" mapstructure:"use_default,omitempty" json:"use_default" temporaljson:"use_default"`

	// ManagedIdentityClientID runs the operation as a specific user-assigned
	// managed identity instead of the VM's system identity.
	ManagedIdentityClientID string `cty:"managed_identity_client_id,optional" hcl:"managed_identity_client_id,optional" mapstructure:"managed_identity_client_id,omitempty" json:"managed_identity_client_id" temporaljson:"managed_identity_client_id"`

	// The fields below authenticate as an app registration in someone else's
	// tenant — a vendor's registry, reached from a control plane that holds no
	// identity there. Managed identity cannot cross a tenant boundary, so this
	// is the only path for that case.
	//
	// They are deliberately unserializable. A Config travels inside plans and
	// Temporal history; ClientSecret and ClientCertificatePEM are long-lived
	// vendor credentials that must not. Resolve them, mint a short-lived
	// registry token, and send that instead (see EnsureACRAuth). The blank
	// tags are what make an accidental send drop the material rather than
	// leak it.
	TenantID             string `cty:"-" hcl:"-" mapstructure:"-" json:"-" temporaljson:"-"`
	ClientID             string `cty:"-" hcl:"-" mapstructure:"-" json:"-" temporaljson:"-"`
	ClientSecret         string `cty:"-" hcl:"-" mapstructure:"-" json:"-" temporaljson:"-"`
	ClientCertificatePEM []byte `cty:"-" hcl:"-" mapstructure:"-" json:"-" temporaljson:"-"`
}

// HasAppRegistrationCredentials reports whether this config carries material
// for a specific app registration, as opposed to relying on whatever ambient
// identity the process happens to have.
func (c Config) HasAppRegistrationCredentials() bool {
	if c.ClientID == "" || c.TenantID == "" {
		return false
	}
	return c.ClientSecret != "" || len(c.ClientCertificatePEM) > 0
}

func (c Config) String() string {
	if c.HasAppRegistrationCredentials() {
		kind := "secret"
		if len(c.ClientCertificatePEM) > 0 {
			kind = "certificate"
		}
		return "app registration " + c.ClientID + " in tenant " + c.TenantID + " (" + kind + ")"
	}
	if c.ManagedIdentityClientID != "" {
		return "user-assigned managed identity " + c.ManagedIdentityClientID
	}
	if c.UseDefault {
		return "default credentials"
	}

	return "managed identity"
}
