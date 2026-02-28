package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/frostyeti/akv/internal/keyvault"
	"github.com/spf13/cobra"
)

type certificateService interface {
	Get(ctx context.Context, name string, version string) (keyvault.CertificateInfo, error)
	Set(ctx context.Context, name string, in keyvault.CertificateCreateInput) error
	Update(ctx context.Context, name string, in keyvault.CertificateUpdateInput) error
	Delete(ctx context.Context, name string) error
	Purge(ctx context.Context, name string) error
}

var certificateServiceFactory = buildCertificateService

func newCertificatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificates",
		Short: "Manage Azure Key Vault certificates",
	}

	cmd.AddCommand(newCertificateGetCmd())
	cmd.AddCommand(newCertificateSetCmd())
	cmd.AddCommand(newCertificateUpdateCmd())
	cmd.AddCommand(newCertificateDeleteCmd())
	cmd.AddCommand(newCertificatePurgeCmd())

	return cmd
}

func newCertificateGetCmd() *cobra.Command {
	version := ""

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get certificate metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
			if err != nil {
				return err
			}

			info, err := service.Get(cmd.Context(), args[0], version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "id=%s contentType=%s\n", info.ID, info.ContentType)
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific certificate version")
	return cmd
}

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

func newCertificateUpdateCmd() *cobra.Command {
	var (
		version    string
		tagValues  []string
		expiresOn  string
		notBefore  string
		enabledSet bool
		enabled    bool
	)

	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update certificate properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
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

			in := keyvault.CertificateUpdateInput{Version: version, Tags: tags, ExpiresOn: expiresAt, NotBefore: notBeforeAt}
			if enabledSet {
				in.Enabled = &enabled
			}

			if err := service.Update(cmd.Context(), args[0], in); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "updated certificate %q\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific certificate version")
	cmd.Flags().StringArrayVar(&tagValues, "tag", nil, "Certificate tag in key=value format")
	cmd.Flags().StringVar(&expiresOn, "expires-on", "", "Expiration time in RFC3339")
	cmd.Flags().StringVar(&notBefore, "not-before", "", "Not-before time in RFC3339")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the certificate is enabled")
	cmd.Flags().BoolVar(&enabledSet, "set-enabled", false, "Apply --enabled value when updating")

	return cmd
}

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

func buildCertificateService(cmd *cobra.Command) (certificateService, error) {
	vaultURL, err := resolveVaultURL(cmd)
	if err != nil {
		return nil, err
	}

	service, err := keyvault.NewCertificatesService(vaultURL)
	if err != nil {
		return nil, err
	}

	return service, nil
}
