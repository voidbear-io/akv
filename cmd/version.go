package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the CLI version string.
var Version = "0.0.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "version",
		Short:  "Print the akv version",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), Version)
			return err
		},
	}
}
