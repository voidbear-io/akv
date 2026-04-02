package cmd

import (
	"errors"
	"fmt"

	"github.com/frostyeti/akv/internal/keyvault"
	"github.com/spf13/cobra"
)

func newSecretEnsureCmd() *cobra.Command {
	opts := secretGenerationOptions{Size: 16}

	cmd := &cobra.Command{
		Use:   "ensure <name>",
		Short: "Ensure a secret exists",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			if _, err := service.Get(cmd.Context(), name, ""); err == nil {
				_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "secret %q already exists\n", name)
				return writeErr
			} else if !errors.Is(err, keyvault.ErrSecretNotFound) {
				return err
			}

			value, err := generateSecretValue(opts)
			if err != nil {
				return err
			}

			if err := service.Set(cmd.Context(), name, value); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "created secret %q\n", name)
			return err
		},
	}

	registerSecretGenerationFlags(cmd, &opts)
	return cmd
}

func newSecretEnsureAliasCmd() *cobra.Command {
	return newSecretEnsureCmd()
}
