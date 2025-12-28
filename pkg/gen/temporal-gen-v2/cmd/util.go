package cmd

func getDir(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "."
}
