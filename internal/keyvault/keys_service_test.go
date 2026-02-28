package keyvault

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azkeys"
)

type fakeKeyClient struct {
	getResp   azkeys.GetKeyResponse
	getErr    error
	createErr error
	updateErr error
	delErr    error
	purgeErr  error

	createCalls int
	updateCalls int
}

func (f *fakeKeyClient) GetKey(ctx context.Context, name string, version string, options *azkeys.GetKeyOptions) (azkeys.GetKeyResponse, error) {
	return f.getResp, f.getErr
}

func (f *fakeKeyClient) CreateKey(ctx context.Context, name string, parameters azkeys.CreateKeyParameters, options *azkeys.CreateKeyOptions) (azkeys.CreateKeyResponse, error) {
	f.createCalls++
	return azkeys.CreateKeyResponse{}, f.createErr
}

func (f *fakeKeyClient) UpdateKey(ctx context.Context, name string, version string, parameters azkeys.UpdateKeyParameters, options *azkeys.UpdateKeyOptions) (azkeys.UpdateKeyResponse, error) {
	f.updateCalls++
	return azkeys.UpdateKeyResponse{}, f.updateErr
}

func (f *fakeKeyClient) DeleteKey(ctx context.Context, name string, options *azkeys.DeleteKeyOptions) (azkeys.DeleteKeyResponse, error) {
	return azkeys.DeleteKeyResponse{}, f.delErr
}

func (f *fakeKeyClient) PurgeDeletedKey(ctx context.Context, name string, options *azkeys.PurgeDeletedKeyOptions) (azkeys.PurgeDeletedKeyResponse, error) {
	return azkeys.PurgeDeletedKeyResponse{}, f.purgeErr
}

func TestGetKeyNotFound(t *testing.T) {
	client := &fakeKeyClient{getErr: &azcore.ResponseError{StatusCode: http.StatusNotFound}}
	svc := NewKeysServiceWithClient(client)

	_, err := svc.Get(context.Background(), "service-key", "")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestSetKey(t *testing.T) {
	client := &fakeKeyClient{}
	svc := NewKeysServiceWithClient(client)

	err := svc.Set(context.Background(), "service-key", KeyCreateInput{Type: "rsa"})
	if err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	if client.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", client.createCalls)
	}
}

func TestUpdateKey(t *testing.T) {
	client := &fakeKeyClient{}
	svc := NewKeysServiceWithClient(client)
	now := time.Now().UTC().Add(6 * time.Hour)
	enabled := true

	err := svc.Update(context.Background(), "service-key", KeyUpdateInput{Version: "v1", Enabled: &enabled, ExpiresOn: &now})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	if client.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", client.updateCalls)
	}
}
