package keyvault

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/frostyeti/akv/internal/auth"
)

var ErrSecretNotFound = errors.New("secret not found")

type secretClient interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
	SetSecret(ctx context.Context, name string, parameters azsecrets.SetSecretParameters, options *azsecrets.SetSecretOptions) (azsecrets.SetSecretResponse, error)
	DeleteSecret(ctx context.Context, name string, options *azsecrets.DeleteSecretOptions) (azsecrets.DeleteSecretResponse, error)
	UpdateSecretProperties(ctx context.Context, name string, version string, parameters azsecrets.UpdateSecretPropertiesParameters, options *azsecrets.UpdateSecretPropertiesOptions) (azsecrets.UpdateSecretPropertiesResponse, error)
	NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse]
	PurgeDeletedSecret(ctx context.Context, name string, options *azsecrets.PurgeDeletedSecretOptions) (azsecrets.PurgeDeletedSecretResponse, error)
}

// SecretUpdateInput stores mutable secret metadata attributes.
type SecretUpdateInput struct {
	Version     string
	ContentType *string
	Enabled     *bool
	ExpiresOn   *time.Time
	NotBefore   *time.Time
	Tags        map[string]string
}

// SecretsService executes secret operations against Azure Key Vault.
type SecretsService struct {
	client secretClient
}

// NewSecretsService constructs a SecretsService for a vault URL.
func NewSecretsService(vaultURL string) (*SecretsService, error) {
	cred, err := auth.NewCredential()
	if err != nil {
		return nil, fmt.Errorf("create Azure credential: %w", err)
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create Azure secrets client: %w", err)
	}

	return &SecretsService{client: client}, nil
}

// NewSecretsServiceWithClient builds a service with an injected client.
func NewSecretsServiceWithClient(client secretClient) *SecretsService {
	return &SecretsService{client: client}
}

// Get fetches a secret value for a secret name and optional version.
func (s *SecretsService) Get(ctx context.Context, name string, version string) (string, error) {
	resp, err := s.client.GetSecret(ctx, name, version, nil)
	if err != nil {
		if isNotFound(err) {
			return "", ErrSecretNotFound
		}

		return "", fmt.Errorf("get secret %q: %w", name, err)
	}

	if resp.Value == nil {
		return "", fmt.Errorf("get secret %q: empty value", name)
	}

	return *resp.Value, nil
}

// Set creates or updates a secret value.
func (s *SecretsService) Set(ctx context.Context, name string, value string) error {
	params := azsecrets.SetSecretParameters{Value: &value}
	_, err := s.client.SetSecret(ctx, name, params, nil)
	if err != nil {
		return fmt.Errorf("set secret %q: %w", name, err)
	}

	return nil
}

// Delete removes a secret.
func (s *SecretsService) Delete(ctx context.Context, name string) error {
	_, err := s.client.DeleteSecret(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrSecretNotFound
		}

		return fmt.Errorf("delete secret %q: %w", name, err)
	}

	return nil
}

// Purge permanently removes a soft-deleted secret.
func (s *SecretsService) Purge(ctx context.Context, name string) error {
	_, err := s.client.PurgeDeletedSecret(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrSecretNotFound
		}

		return fmt.Errorf("purge secret %q: %w", name, err)
	}

	return nil
}

// Update updates mutable properties for a secret version.
func (s *SecretsService) Update(ctx context.Context, name string, in SecretUpdateInput) error {
	params := azsecrets.UpdateSecretPropertiesParameters{}

	if in.ContentType != nil {
		params.ContentType = in.ContentType
	}

	if in.Tags != nil {
		params.Tags = make(map[string]*string, len(in.Tags))
		for k, v := range in.Tags {
			val := v
			params.Tags[k] = &val
		}
	}

	if in.Enabled != nil || in.ExpiresOn != nil || in.NotBefore != nil {
		params.SecretAttributes = &azsecrets.SecretAttributes{}
		if in.Enabled != nil {
			params.SecretAttributes.Enabled = in.Enabled
		}
		if in.ExpiresOn != nil {
			params.SecretAttributes.Expires = in.ExpiresOn
		}
		if in.NotBefore != nil {
			params.SecretAttributes.NotBefore = in.NotBefore
		}
	}

	_, err := s.client.UpdateSecretProperties(ctx, name, in.Version, params, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrSecretNotFound
		}

		return fmt.Errorf("update secret %q: %w", name, err)
	}

	return nil
}

// List retrieves all secret names from the vault.
func (s *SecretsService) List(ctx context.Context) ([]string, error) {
	var secrets []string

	pager := s.client.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list secrets: %w", err)
		}

		for _, secret := range page.Value {
			if secret.ID != nil {
				secrets = append(secrets, secret.ID.Name())
			}
		}
	}

	return secrets, nil
}

func isNotFound(err error) bool {
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return respErr.StatusCode == http.StatusNotFound
	}

	return false
}
