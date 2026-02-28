package keyvault

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type fakeSecretClient struct {
	getResp  azsecrets.GetSecretResponse
	getErr   error
	setErr   error
	delErr   error
	updErr   error
	purgeErr error

	setCalls   int
	delCalls   int
	updCalls   int
	purgeCalls int
}

func (f *fakeSecretClient) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return f.getResp, f.getErr
}

func (f *fakeSecretClient) SetSecret(ctx context.Context, name string, parameters azsecrets.SetSecretParameters, options *azsecrets.SetSecretOptions) (azsecrets.SetSecretResponse, error) {
	f.setCalls++
	return azsecrets.SetSecretResponse{}, f.setErr
}

func (f *fakeSecretClient) DeleteSecret(ctx context.Context, name string, options *azsecrets.DeleteSecretOptions) (azsecrets.DeleteSecretResponse, error) {
	f.delCalls++
	return azsecrets.DeleteSecretResponse{}, f.delErr
}

func (f *fakeSecretClient) UpdateSecretProperties(ctx context.Context, name string, version string, parameters azsecrets.UpdateSecretPropertiesParameters, options *azsecrets.UpdateSecretPropertiesOptions) (azsecrets.UpdateSecretPropertiesResponse, error) {
	f.updCalls++
	return azsecrets.UpdateSecretPropertiesResponse{}, f.updErr
}

func (f *fakeSecretClient) NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	// Return a pager that has no pages - for testing, we'll just return nil
	// This is a simplified implementation that won't be called in existing tests
	return nil
}

func (f *fakeSecretClient) PurgeDeletedSecret(ctx context.Context, name string, options *azsecrets.PurgeDeletedSecretOptions) (azsecrets.PurgeDeletedSecretResponse, error) {
	f.purgeCalls++
	return azsecrets.PurgeDeletedSecretResponse{}, f.purgeErr
}

func TestGetSecretNotFound(t *testing.T) {
	client := &fakeSecretClient{
		getErr: &azcore.ResponseError{StatusCode: http.StatusNotFound},
	}

	svc := NewSecretsServiceWithClient(client)
	_, err := svc.Get(context.Background(), "db-password", "")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Fatalf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestSetSecret(t *testing.T) {
	client := &fakeSecretClient{}
	svc := NewSecretsServiceWithClient(client)

	if err := svc.Set(context.Background(), "db-password", "secret"); err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	if client.setCalls != 1 {
		t.Fatalf("expected one set call, got %d", client.setCalls)
	}
}

func TestDeleteSecretNotFound(t *testing.T) {
	client := &fakeSecretClient{
		delErr: &azcore.ResponseError{StatusCode: http.StatusNotFound},
	}
	svc := NewSecretsServiceWithClient(client)

	err := svc.Delete(context.Background(), "db-password")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Fatalf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestUpdateSecret(t *testing.T) {
	client := &fakeSecretClient{}
	svc := NewSecretsServiceWithClient(client)

	expires := time.Now().UTC().Add(24 * time.Hour)
	enabled := true
	contentType := "text/plain"

	err := svc.Update(context.Background(), "db-password", SecretUpdateInput{
		Version:     "1234",
		ContentType: &contentType,
		Enabled:     &enabled,
		ExpiresOn:   &expires,
		Tags: map[string]string{
			"team": "platform",
		},
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	if client.updCalls != 1 {
		t.Fatalf("expected one update call, got %d", client.updCalls)
	}
}
