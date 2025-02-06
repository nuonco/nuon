package cfngen

// CfgFile represents the temporary prototype configuration file we're using to iterate on generator inputs and config values.
// Remove this once that work is done.
type CfgFile struct {
	Internal  InternalValues  `toml:"internal"`
	VendorCfg AppConfigValues `toml:"vendorcfg"`
}
