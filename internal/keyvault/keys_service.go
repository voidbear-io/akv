package keyvault

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys"
	"github.com/voidbear-io/akv/internal/auth"
)

var ErrKeyNotFound = errors.New("key not found")

type keyClient interface {
	GetKey(ctx context.Context, name string, version string, options *azkeys.GetKeyOptions) (azkeys.GetKeyResponse, error)
	CreateKey(ctx context.Context, name string, parameters azkeys.CreateKeyParameters, options *azkeys.CreateKeyOptions) (azkeys.CreateKeyResponse, error)
	NewListKeyPropertiesPager(options *azkeys.ListKeyPropertiesOptions) *runtime.Pager[azkeys.ListKeyPropertiesResponse]
	UpdateKey(ctx context.Context, name string, version string, parameters azkeys.UpdateKeyParameters, options *azkeys.UpdateKeyOptions) (azkeys.UpdateKeyResponse, error)
	DeleteKey(ctx context.Context, name string, options *azkeys.DeleteKeyOptions) (azkeys.DeleteKeyResponse, error)
	PurgeDeletedKey(ctx context.Context, name string, options *azkeys.PurgeDeletedKeyOptions) (azkeys.PurgeDeletedKeyResponse, error)
}

// KeyInfo stores selected key metadata for command output.
type KeyInfo struct {
	ID   string
	Type string
}

// KeyCreateInput contains key creation parameters.
type KeyCreateInput struct {
	Type string
	Tags map[string]string
}

// KeyUpdateInput contains mutable key metadata fields.
type KeyUpdateInput struct {
	Version   string
	Enabled   *bool
	ExpiresOn *time.Time
	NotBefore *time.Time
	Tags      map[string]string
}

// KeysService executes key operations against Azure Key Vault.
type KeysService struct {
	client keyClient
}

// NewKeysService constructs a KeysService for a vault URL.
func NewKeysService(vaultURL string) (*KeysService, error) {
	cred, err := auth.NewCredential()
	if err != nil {
		return nil, fmt.Errorf("create Azure credential: %w", err)
	}

	client, err := azkeys.NewClient(vaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create Azure keys client: %w", err)
	}

	return &KeysService{client: client}, nil
}

// NewKeysServiceWithClient builds a service with an injected client.
func NewKeysServiceWithClient(client keyClient) *KeysService {
	return &KeysService{client: client}
}

// Get fetches key metadata by key name and optional version.
func (s *KeysService) Get(ctx context.Context, name string, version string) (KeyInfo, error) {
	resp, err := s.client.GetKey(ctx, name, version, nil)
	if err != nil {
		if isNotFound(err) {
			return KeyInfo{}, ErrKeyNotFound
		}

		return KeyInfo{}, fmt.Errorf("get key %q: %w", name, err)
	}

	info := KeyInfo{}
	if resp.Key != nil {
		if resp.Key.KID != nil {
			info.ID = string(*resp.Key.KID)
		}
		if resp.Key.Kty != nil {
			info.Type = string(*resp.Key.Kty)
		}
	}

	return info, nil
}

// Set creates a new key version.
func (s *KeysService) Set(ctx context.Context, name string, in KeyCreateInput) error {
	kty, err := parseKeyType(in.Type)
	if err != nil {
		return err
	}

	params := azkeys.CreateKeyParameters{Kty: &kty}
	if in.Tags != nil {
		params.Tags = make(map[string]*string, len(in.Tags))
		for k, v := range in.Tags {
			val := v
			params.Tags[k] = &val
		}
	}

	_, err = s.client.CreateKey(ctx, name, params, nil)
	if err != nil {
		return fmt.Errorf("set key %q: %w", name, err)
	}

	return nil
}

// List retrieves all key names from the vault.
func (s *KeysService) List(ctx context.Context) ([]string, error) {
	var keys []string

	pager := s.client.NewListKeyPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list keys: %w", err)
		}

		for _, key := range page.Value {
			if key.KID != nil {
				keys = append(keys, key.KID.Name())
			}
		}
	}

	return keys, nil
}

// Update updates mutable key properties for an optional version.
func (s *KeysService) Update(ctx context.Context, name string, in KeyUpdateInput) error {
	params := azkeys.UpdateKeyParameters{}
	if in.Tags != nil {
		params.Tags = make(map[string]*string, len(in.Tags))
		for k, v := range in.Tags {
			val := v
			params.Tags[k] = &val
		}
	}

	if in.Enabled != nil || in.ExpiresOn != nil || in.NotBefore != nil {
		params.KeyAttributes = &azkeys.KeyAttributes{}
		if in.Enabled != nil {
			params.KeyAttributes.Enabled = in.Enabled
		}
		if in.ExpiresOn != nil {
			params.KeyAttributes.Expires = in.ExpiresOn
		}
		if in.NotBefore != nil {
			params.KeyAttributes.NotBefore = in.NotBefore
		}
	}

	_, err := s.client.UpdateKey(ctx, name, in.Version, params, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrKeyNotFound
		}

		return fmt.Errorf("update key %q: %w", name, err)
	}

	return nil
}

// Delete removes a key.
func (s *KeysService) Delete(ctx context.Context, name string) error {
	_, err := s.client.DeleteKey(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrKeyNotFound
		}

		return fmt.Errorf("delete key %q: %w", name, err)
	}

	return nil
}

// Purge permanently removes a soft-deleted key.
func (s *KeysService) Purge(ctx context.Context, name string) error {
	_, err := s.client.PurgeDeletedKey(ctx, name, nil)
	if err != nil {
		if isNotFound(err) {
			return ErrKeyNotFound
		}

		return fmt.Errorf("purge key %q: %w", name, err)
	}

	return nil
}

func parseKeyType(v string) (azkeys.KeyType, error) {
	switch v {
	case "", "rsa", "RSA", "Rsa":
		return azkeys.KeyTypeRSA, nil
	case "ec", "EC", "Ec":
		return azkeys.KeyTypeEC, nil
	case "oct", "OCT", "Oct":
		return azkeys.KeyTypeOct, nil
	case "rsa-hsm", "RSA-HSM", "RsaHsm":
		return azkeys.KeyTypeRSAHSM, nil
	case "ec-hsm", "EC-HSM", "EcHsm":
		return azkeys.KeyTypeECHSM, nil
	case "oct-hsm", "OCT-HSM", "OctHsm":
		return azkeys.KeyTypeOctHSM, nil
	default:
		return "", fmt.Errorf("unsupported key type %q", v)
	}
}
