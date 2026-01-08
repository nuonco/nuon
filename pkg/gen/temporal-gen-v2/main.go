package main

import (
	"fmt"
	"os"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
