package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newCertificateSetCmd() *cobra.Command {
	subject := ""
	tags := []string{}

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Create or import a certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
			if err != nil {
				return err
			}

			tagMap, err := parseTags(tags)
			if err != nil {
				return err
			}

			if err := service.Set(cmd.Context(), args[0], keyvault.CertificateCreateInput{Subject: subject, Tags: tagMap}); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "set certificate %q\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Certificate subject; default CN=<name>")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Certificate tag in key=value format")
	return cmd
}
