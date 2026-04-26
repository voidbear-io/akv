package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

type certificateService interface {
	Get(ctx context.Context, name string, version string) (keyvault.CertificateInfo, error)
	Set(ctx context.Context, name string, in keyvault.CertificateCreateInput) error
	ImportCertificate(ctx context.Context, name string, in keyvault.CertificateImportInput) error
	List(ctx context.Context) ([]string, error)
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
	cmd.AddCommand(newCertificateListCmd())
	cmd.AddCommand(newCertificateDownloadCmd())
	cmd.AddCommand(newCertificateUploadCmd())
	cmd.AddCommand(newCertificateSetCmd())
	cmd.AddCommand(newCertificateUpdateCmd())
	cmd.AddCommand(newCertificateDeleteCmd())
	cmd.AddCommand(newCertificatePurgeCmd())

	return cmd
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
