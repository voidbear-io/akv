package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newKeySetCmd() *cobra.Command {
	keyType := "rsa"
	tags := []string{}

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Create or update a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := keyServiceFactory(cmd)
			if err != nil {
				return err
			}

			tagMap, err := parseTags(tags)
			if err != nil {
				return err
			}

			if err := service.Set(cmd.Context(), args[0], keyvault.KeyCreateInput{Type: keyType, Tags: tagMap}); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "set key %q\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVar(&keyType, "type", "rsa", "Key type: rsa|ec|oct|rsa-hsm|ec-hsm|oct-hsm")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Key tag in key=value format")
	return cmd
}
