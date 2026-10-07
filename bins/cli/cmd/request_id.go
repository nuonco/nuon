package cmd

import "github.com/spf13/cobra"

const requestIDFlagHelp = "Idempotency key for this call. A retry with the same key and the same arguments returns the original result. A different request returns 409. Max 255 characters."

func addRequestIDFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVar(dest, "request-id", "", requestIDFlagHelp)
}
