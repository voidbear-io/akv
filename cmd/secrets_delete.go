package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newSecretDeleteCmd() *cobra.Command {
	var purge bool

	cmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			err = service.Delete(cmd.Context(), name)
			if err != nil && !errors.Is(err, keyvault.ErrSecretNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrSecretNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "secret %q not found\n", name)
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "deleted secret %q\n", name)
			if err != nil {
				return err
			}

			if purge {
				err = service.Purge(cmd.Context(), name)
				if err != nil && !errors.Is(err, keyvault.ErrSecretNotFound) {
					return err
				}

				if errors.Is(err, keyvault.ErrSecretNotFound) {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "secret %q not found for purge\n", name)
					return err
				}

				_, err = fmt.Fprintf(cmd.OutOrStdout(), "purged secret %q\n", name)
			}

			return err
		},
	}

	cmd.Flags().BoolVar(&purge, "purge", false, "Also purge the secret after deletion (permanent removal)")
	return cmd
}

func newSecretDeleteAliasCmd() *cobra.Command {
	return newSecretDeleteCmd()
}

func newSecretPurgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge <name>",
		Short: "Purge a deleted secret",
		Long:  "Permanently removes a soft-deleted secret from Azure Key Vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			err = service.Purge(cmd.Context(), name)
			if err != nil && !errors.Is(err, keyvault.ErrSecretNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrSecretNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "secret %q not found\n", name)
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "purged secret %q\n", name)
			return err
		},
	}

	return cmd
}
