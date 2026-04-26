package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newSecretUpdateCmd() *cobra.Command {
	var (
		version     string
		contentType string
		tagValues   []string
		expiresOn   string
		notBefore   string
		enabledSet  bool
		enabled     bool
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update secret properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			tags, err := parseTags(tagValues)
			if err != nil {
				return err
			}

			expiresAt, err := parseOptionalTime(expiresOn)
			if err != nil {
				return err
			}

			notBeforeAt, err := parseOptionalTime(notBefore)
			if err != nil {
				return err
			}

			in := keyvault.SecretUpdateInput{Version: version, Tags: tags, ExpiresOn: expiresAt, NotBefore: notBeforeAt}
			if contentType != "" {
				in.ContentType = &contentType
			}
			if enabledSet {
				in.Enabled = &enabled
			}

			if err := service.Update(cmd.Context(), args[0], in); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "updated secret %q\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific secret version GUID")
	cmd.Flags().StringVar(&contentType, "content-type", "", "Secret content type")
	cmd.Flags().StringArrayVar(&tagValues, "tag", nil, "Secret tag in key=value format")
	cmd.Flags().StringVar(&expiresOn, "expires-on", "", "Expiration time in RFC3339")
	cmd.Flags().StringVar(&notBefore, "not-before", "", "Not-before time in RFC3339")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the secret is enabled")
	cmd.Flags().BoolVar(&enabledSet, "set-enabled", false, "Apply --enabled value when updating")

	return cmd
}
