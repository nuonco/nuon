package sandboxes

type VPNOutputs struct {
	Name      string        `mapstructure:"name" validate:"required"`
	SubnetIDs []interface{} `mapstructure:"subnet_ids" validate:"required" faker:"stringSliceAsInt"`
}

type AccountOutputs struct {
	SubscriptionID string `mapstructure:"subscription_id" validate:"required"`
	Location       string `mapstructure:"location" validate:"required"`
}

type ACROutputs struct {
	ID          string `mapstructure:"id" validate:"required"`
	LoginServer string `mapstructure:"login_server" validate:"required"`
	TokenID     string `mapstructure:"token_id" validate:"required"`
	Password    string `mapstructure:"password" validate:"required"`
}
