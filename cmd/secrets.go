package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/frostyeti/akv/internal/config"
	"github.com/frostyeti/akv/internal/keyvault"
	"github.com/spf13/cobra"
)

type secretService interface {
	Get(ctx context.Context, name string, version string) (string, error)
	Set(ctx context.Context, name string, value string) error
	Delete(ctx context.Context, name string) error
	Update(ctx context.Context, name string, in keyvault.SecretUpdateInput) error
	List(ctx context.Context) ([]string, error)
	Purge(ctx context.Context, name string) error
}

var secretServiceFactory = buildSecretService

func newSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage Azure Key Vault secrets",
	}

	cmd.AddCommand(newSecretGetCmd("get", "Get a secret value"))
	cmd.AddCommand(newSecretSetCmd("set", "Set a secret value"))
	cmd.AddCommand(newSecretDeleteCmd("rm", "Delete a secret"))
	cmd.AddCommand(newSecretPurgeCmd("purge", "Purge a deleted secret"))
	cmd.AddCommand(newSecretEnsureCmd("ensure", "Ensure a secret exists"))
	cmd.AddCommand(newSecretUpdateCmd("update", "Update secret properties"))
	cmd.AddCommand(newSecretListCmd("ls", "List secrets"))

	return cmd
}

func newSecretRootAliasCmds() []*cobra.Command {
	return []*cobra.Command{
		newSecretGetCmd("get", "Alias for secrets get"),
		newSecretSetCmd("set", "Alias for secrets set"),
		newSecretDeleteCmd("rm", "Alias for secrets rm"),
		newSecretEnsureCmd("ensure", "Alias for secrets ensure"),
	}
}

func newSecretGetCmd(use string, short string) *cobra.Command {
	version := ""

	cmd := &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			value, err := service.Get(cmd.Context(), args[0], version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific secret version GUID")

	return cmd
}

func newSecretSetCmd(use string, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <name> <value>",
		Short: short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			if err := service.Set(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "set secret %q\n", args[0])
			return err
		},
	}
}

func newSecretDeleteCmd(use string, short string) *cobra.Command {
	var purge bool

	cmd := &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			err = service.Delete(cmd.Context(), args[0])
			if err != nil && !errors.Is(err, keyvault.ErrSecretNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrSecretNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "secret %q not found\n", args[0])
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "deleted secret %q\n", args[0])
			if err != nil {
				return err
			}

			// If purge flag is set, also purge the secret
			if purge {
				err = service.Purge(cmd.Context(), args[0])
				if err != nil && !errors.Is(err, keyvault.ErrSecretNotFound) {
					return err
				}

				if errors.Is(err, keyvault.ErrSecretNotFound) {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "secret %q not found for purge\n", args[0])
					return err
				}

				_, err = fmt.Fprintf(cmd.OutOrStdout(), "purged secret %q\n", args[0])
			}

			return err
		},
	}

	cmd.Flags().BoolVar(&purge, "purge", false, "Also purge the secret after deletion (permanent removal)")

	return cmd
}

func newSecretPurgeCmd(use string, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Long:  "Permanently removes a soft-deleted secret from Azure Key Vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			err = service.Purge(cmd.Context(), args[0])
			if err != nil && !errors.Is(err, keyvault.ErrSecretNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrSecretNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "secret %q not found\n", args[0])
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "purged secret %q\n", args[0])
			return err
		},
	}
}

func newSecretEnsureCmd(use string, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <name> <value>",
		Short: short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			_, err = service.Get(cmd.Context(), args[0], "")
			if err == nil {
				_, writeErr := fmt.Fprintf(cmd.OutOrStdout(), "secret %q already exists\n", args[0])
				return writeErr
			}

			if !errors.Is(err, keyvault.ErrSecretNotFound) {
				return err
			}

			if err := service.Set(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "created secret %q\n", args[0])
			return err
		},
	}
}

func newSecretUpdateCmd(use string, short string) *cobra.Command {
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
		Use:   use + " <name>",
		Short: short,
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

func newSecretListCmd(use string, short string) *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   use + " [pattern]",
		Short: short,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				pattern = args[0]
			}

			// Placeholder - actual implementation would list secrets from vault
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listing secrets (pattern: %s)\n", pattern)
			return nil
		},
	}

	cmd.Flags().StringVarP(&pattern, "filter", "f", "", "Glob pattern to filter secrets")

	return cmd
}

func buildSecretService(cmd *cobra.Command) (secretService, error) {
	vaultURL, err := resolveVaultURL(cmd)
	if err != nil {
		return nil, err
	}

	service, err := keyvault.NewSecretsService(vaultURL)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func resolveVaultURL(cmd *cobra.Command) (string, error) {
	// Check command line flag first
	vaultURL, err := cmd.Root().PersistentFlags().GetString("vault-url")
	if err != nil {
		return "", err
	}

	// Then check environment variable
	if vaultURL == "" {
		vaultURL = os.Getenv("AKV_VAULT_URL")
	}

	// Finally check config for current vault
	if vaultURL == "" {
		mgr, err := config.NewManager()
		if err == nil {
			url, err := mgr.GetVaultURL("")
			if err == nil {
				vaultURL = url
			}
		}
	}

	if vaultURL == "" {
		return "", ErrVaultURLRequired
	}

	return vaultURL, nil
}
