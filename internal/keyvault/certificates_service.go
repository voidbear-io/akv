package keyvault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
	"github.com/voidbear-io/akv/internal/auth"
)

var ErrCertificateNotFound = errors.New("certificate not found")

type certificateClient interface {
	GetCertificate(ctx context.Context, name string, version string, options *azcertificates.GetCertificateOptions) (azcertificates.GetCertificateResponse, error)
	CreateCertificate(ctx context.Context, name string, parameters azcertificates.CreateCertificateParameters, options *azcertificates.CreateCertificateOptions) (azcertificates.CreateCertificateResponse, error)
	ImportCertificate(ctx context.Context, name string, parameters azcertificates.ImportCertificateParameters, options *azcertificates.ImportCertificateOptions) (azcertificates.ImportCertificateResponse, error)
	NewListCertificatePropertiesPager(options *azcertificates.ListCertificatePropertiesOptions) *runtime.Pager[azcertificates.ListCertificatePropertiesResponse]
	UpdateCertificate(ctx context.Context, name string, version string, parameters azcertificates.UpdateCertificateParameters, options *azcertificates.UpdateCertificateOptions) (azcertificates.UpdateCertificateResponse, error)
	DeleteCertificate(ctx context.Context, name string, options *azcertificates.DeleteCertificateOptions) (azcertificates.DeleteCertificateResponse, error)
	PurgeDeletedCertificate(ctx context.Context, name string, options *azcertificates.PurgeDeletedCertificateOptions) (azcertificates.PurgeDeletedCertificateResponse, error)
}

// CertificateInfo stores selected certificate metadata for command output.
type CertificateInfo struct {
	ID          string
	ContentType string
	SID         string
}

// CertificateImportInput contains certificate import parameters.
type CertificateImportInput struct {
	Base64EncodedCertificate string
	Password                 string
	Tags                     map[string]string
}

// CertificateCreateInput contains certificate creation parameters.
type CertificateCreateInput struct {
	Subject string
	Tags    map[string]string
}

// CertificateUpdateInput contains mutable certificate metadata fields.
type CertificateUpdateInput struct {
	Version   string
	Enabled   *bool
	ExpiresOn *time.Time
	NotBefore *time.Time
	Tags      map[string]string
}

// CertificatesService executes certificate operations against Azure Key Vault.
type CertificatesService struct {
	client certificateClient
}

// NewCertificatesService constructs a CertificatesService for a vault URL.
func NewCertificatesService(vaultURL string) (*CertificatesService, error) {
	cred, err := auth.NewCredential()
	if err != nil {
		return nil, fmt.Errorf("create Azure credential: %w", err)
	}

	client, err := azcertificates.NewClient(vaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create Azure certificates client: %w", err)
	}

	return &CertificatesService{client: client}, nil
}

// NewCertificatesServiceWithClient builds a service with an injected client.
func NewCertificatesServiceWithClient(client certificateClient) *CertificatesService {
	return &CertificatesService{client: client}
}

// Get fetches certificate metadata by name and optional version.
func (s *CertificatesService) Get(ctx context.Context, name string, version string) (CertificateInfo, error) {
	resp, err := s.client.GetCertificate(ctx, name, version, nil)
	if err != nil {
		if isNotFound(err) {
			return CertificateInfo{}, ErrCertificateNotFound
		}

		return CertificateInfo{}, fmt.Errorf("get certificate %q: %w", name, err)
	}

	info := CertificateInfo{}
	if resp.ID != nil {
		info.ID = string(*resp.ID)
	}
	if resp.ContentType != nil {
		info.ContentType = *resp.ContentType
	}
	if resp.SID != nil {
		info.SID = string(*resp.SID)
	}

	return info, nil
}

// Import imports a certificate file into Key Vault.
func (s *CertificatesService) Import(ctx context.Context, name string, in CertificateImportInput) error {
	params := azcertificates.ImportCertificateParameters{Base64EncodedCertificate: &in.Base64EncodedCertificate}
	if in.Password != "" {
		params.Password = &in.Password
	}
	if in.Tags != nil {
		params.Tags = make(map[string]*string, len(in.Tags))
		for k, v := range in.Tags {
			val := v
			params.Tags[k] = &val
		}
	}

	_, err := s.client.ImportCertificate(ctx, name, params, nil)
	if err != nil {
		return fmt.Errorf("import certificate %q: %w", name, err)
	}

	return nil
}

// ImportCertificate imports a certificate file into Key Vault.
func (s *CertificatesService) ImportCertificate(ctx context.Context, name string, in CertificateImportInput) error {
	return s.Import(ctx, name, in)
}

// Set creates a certificate with a default self-signed policy.
func (s *CertificatesService) Set(ctx context.Context, name string, in CertificateCreateInput) error {
	if in.Subject == "" {
		in.Subject = "CN=" + name
	}

	params := azcertificates.CreateCertificateParameters{
		CertificatePolicy: &azcertificates.CertificatePolicy{
			IssuerParameters: &azcertificates.IssuerParameters{Name: to.Ptr("Self")},
			X509CertificateProperties: &azcertificates.X509CertificateProperties{
				Subject: &in.Subject,
			},
		},
	}

	if in.Tags != nil {
		params.Tags = make(map[string]*string, len(in.Tags))
		for k, v := range in.Tags {
			val := v
			params.Tags[k] = &val
		}
	}

	_, err := s.client.CreateCertificate(ctx, name, params, nil)
	if err != nil {
		return fmt.Errorf("set certificate %q: %w", name, err)
	}

	return nil
}

// List retrieves all certificate names from the vault.
func (s *CertificatesService) List(ctx context.Context) ([]string, error) {
	var certs []string

	pager := s.client.NewListCertificatePropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list certificates: %w", err)
		}

		for _, cert := range page.Value {
			if cert.ID != nil {
				certs = append(certs, cert.ID.Name())
			}
		}
	}

	return certs, nil
}

// Update updates mutable certificate properties for an optional version.
func (s *CertificatesService) Update(ctx context.Context, name string, in CertificateUpdateInput) error {
	params := azcertificates.UpdateCertificateParameters{}

	if in.Tags != nil {
		params.Tags = make(map[string]*string, len(in.Tags))
		for k, v := range in.Tags {
			val := v
			params.Tags[k] = &val
		}
	}

	if in.Enabled != nil || in.ExpiresOn != nil || in.NotBefore != nil {
		params.CertificateAttributes = &azcertificates.CertificateAttributes{}
		if in.Enabled != nil {
			params.CertificateAttributes.Enabled = in.Enabled
		}
		if in.ExpiresOn != nil {
			params.CertificateAttributes.Expires = in.ExpiresOn
		}
		if in.NotBefore != nil {
			params.CertificateAttributes.NotBefore = in.NotBefore
		}
	}

	_, err := s.client.UpdateCertificate(ctx, name, in.Version, params, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrCertificateNotFound
		}

		return fmt.Errorf("update certificate %q: %w", name, err)
	}

	return nil
}

// Delete removes a certificate.
func (s *CertificatesService) Delete(ctx context.Context, name string) error {
	_, err := s.client.DeleteCertificate(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrCertificateNotFound
		}

		return fmt.Errorf("delete certificate %q: %w", name, err)
	}

	return nil
}

// Purge permanently removes a soft-deleted certificate.
func (s *CertificatesService) Purge(ctx context.Context, name string) error {
	_, err := s.client.PurgeDeletedCertificate(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrCertificateNotFound
		}

		return fmt.Errorf("purge certificate %q: %w", name, err)
	}

	return nil
}
