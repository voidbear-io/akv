package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newCertificatePurgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "purge <name>",
		Short: "Permanently purge a deleted certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
			if err != nil {
				return err
			}

			err = service.Purge(cmd.Context(), args[0])
			if err != nil && !errors.Is(err, keyvault.ErrCertificateNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrCertificateNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "certificate %q not found\n", args[0])
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "purged certificate %q\n", args[0])
			return err
		},
	}
}
