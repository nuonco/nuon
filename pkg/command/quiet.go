package command

var quietMode bool

// SetQuietMode is a global variable that will control the entire output of this package until set otherwise.
// If set to true, we write all logs to io.Discard, instead of stderr/stdout.
func SetQuietMode(val bool) {
	quietMode = val
}
