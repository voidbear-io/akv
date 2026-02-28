package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/frostyeti/akv/internal/keyvault"
	"github.com/spf13/cobra"
)

type keyService interface {
	Get(ctx context.Context, name string, version string) (keyvault.KeyInfo, error)
	Set(ctx context.Context, name string, in keyvault.KeyCreateInput) error
	Update(ctx context.Context, name string, in keyvault.KeyUpdateInput) error
	Delete(ctx context.Context, name string) error
	Purge(ctx context.Context, name string) error
}

var keyServiceFactory = buildKeyService

func newKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage Azure Key Vault keys",
	}

	cmd.AddCommand(newKeyGetCmd())
	cmd.AddCommand(newKeySetCmd())
	cmd.AddCommand(newKeyUpdateCmd())
	cmd.AddCommand(newKeyDeleteCmd())
	cmd.AddCommand(newKeyPurgeCmd())

	return cmd
}

func newKeyGetCmd() *cobra.Command {
	version := ""
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get key metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := keyServiceFactory(cmd)
			if err != nil {
				return err
			}

			info, err := service.Get(cmd.Context(), args[0], version)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "id=%s type=%s\n", info.ID, info.Type)
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific key version")
	return cmd
}

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

func newKeyUpdateCmd() *cobra.Command {
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
		Short: "Update key properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := keyServiceFactory(cmd)
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

			in := keyvault.KeyUpdateInput{Version: version, Tags: tags, ExpiresOn: expiresAt, NotBefore: notBeforeAt}
			if enabledSet {
				in.Enabled = &enabled
			}

			if err := service.Update(cmd.Context(), args[0], in); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "updated key %q\n", args[0])
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific key version")
	cmd.Flags().StringArrayVar(&tagValues, "tag", nil, "Key tag in key=value format")
	cmd.Flags().StringVar(&expiresOn, "expires-on", "", "Expiration time in RFC3339")
	cmd.Flags().StringVar(&notBefore, "not-before", "", "Not-before time in RFC3339")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Whether the key is enabled")
	cmd.Flags().BoolVar(&enabledSet, "set-enabled", false, "Apply --enabled value when updating")

	return cmd
}

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

func newKeyPurgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "purge <name>",
		Short: "Permanently purge a deleted key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := keyServiceFactory(cmd)
			if err != nil {
				return err
			}

			err = service.Purge(cmd.Context(), args[0])
			if err != nil && !errors.Is(err, keyvault.ErrKeyNotFound) {
				return err
			}

			if errors.Is(err, keyvault.ErrKeyNotFound) {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "key %q not found\n", args[0])
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "purged key %q\n", args[0])
			return err
		},
	}
}

func buildKeyService(cmd *cobra.Command) (keyService, error) {
	vaultURL, err := resolveVaultURL(cmd)
	if err != nil {
		return nil, err
	}

	service, err := keyvault.NewKeysService(vaultURL)
	if err != nil {
		return nil, err
	}

	return service, nil
}
