package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/config"
)

func newUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the current vault",
		Long:  "Set the current vault used by akv commands.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			name := args[0]
			if vault := mgr.GetVault(name); vault == nil {
				return fmt.Errorf("vault %q not found", name)
			}

			if err := mgr.SetCurrentVault(name); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "using vault %q\n", name)
			return err
		},
	}
}
