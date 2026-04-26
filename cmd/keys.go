package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

type keyService interface {
	Get(ctx context.Context, name string, version string) (keyvault.KeyInfo, error)
	Set(ctx context.Context, name string, in keyvault.KeyCreateInput) error
	List(ctx context.Context) ([]string, error)
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
	cmd.AddCommand(newKeyListCmd())
	cmd.AddCommand(newKeySetCmd())
	cmd.AddCommand(newKeyUpdateCmd())
	cmd.AddCommand(newKeyDeleteCmd())
	cmd.AddCommand(newKeyPurgeCmd())

	return cmd
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
