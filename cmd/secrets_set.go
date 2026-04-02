package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSecretSetCmd() *cobra.Command {
	opts := secretGenerationOptions{Size: 16}

	cmd := &cobra.Command{
		Use:   "set <name> [value]",
		Short: "Set a secret value",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			value := ""
			if len(args) == 2 {
				value = args[1]
			}

			if value != "" && opts.Generate {
				return fmt.Errorf("value and --generate are mutually exclusive")
			}

			if value == "" {
				generated, err := generateSecretValue(opts)
				if err != nil {
					return err
				}
				value = generated
			}

			if err := service.Set(cmd.Context(), name, value); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "set secret %q\n", name)
			return err
		},
	}

	registerSecretGenerationFlags(cmd, &opts)
	return cmd
}

func newSecretSetAliasCmd() *cobra.Command {
	return newSecretSetCmd()
}
