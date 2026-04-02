package cmd

import "github.com/spf13/cobra"

func newLsCmd() *cobra.Command {
	return newSecretListCmd()
}
