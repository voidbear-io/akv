package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newKeyDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := keyServiceFactory(cmd)
			if err != nil {
				return err
			}

			err = service.Delete(cmd.Context(), args[0])
			if err != nil && !errors.Is(err, keyvault.ErrKeyNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrKeyNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "key %q not found\n", args[0])
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "deleted key %q\n", args[0])
			return err
		},
	}
}
