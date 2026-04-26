package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newCertificateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
			if err != nil {
				return err
			}

			err = service.Delete(cmd.Context(), args[0])
			if err != nil && !errors.Is(err, keyvault.ErrCertificateNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrCertificateNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "certificate %q not found\n", args[0])
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "deleted certificate %q\n", args[0])
			return err
		},
	}
}
