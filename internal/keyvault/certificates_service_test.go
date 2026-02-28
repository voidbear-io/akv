package keyvault

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
)

type fakeCertificateClient struct {
	getResp    azcertificates.GetCertificateResponse
	getErr     error
	createErr  error
	updateErr  error
	deleteErr  error
	purgeErr   error
	createCall int
	updateCall int
}

func (f *fakeCertificateClient) GetCertificate(ctx context.Context, name string, version string, options *azcertificates.GetCertificateOptions) (azcertificates.GetCertificateResponse, error) {
	return f.getResp, f.getErr
}

func (f *fakeCertificateClient) CreateCertificate(ctx context.Context, name string, parameters azcertificates.CreateCertificateParameters, options *azcertificates.CreateCertificateOptions) (azcertificates.CreateCertificateResponse, error) {
	f.createCall++
	return azcertificates.CreateCertificateResponse{}, f.createErr
}

func (f *fakeCertificateClient) UpdateCertificate(ctx context.Context, name string, version string, parameters azcertificates.UpdateCertificateParameters, options *azcertificates.UpdateCertificateOptions) (azcertificates.UpdateCertificateResponse, error) {
	f.updateCall++
	return azcertificates.UpdateCertificateResponse{}, f.updateErr
}

func (f *fakeCertificateClient) DeleteCertificate(ctx context.Context, name string, options *azcertificates.DeleteCertificateOptions) (azcertificates.DeleteCertificateResponse, error) {
	return azcertificates.DeleteCertificateResponse{}, f.deleteErr
}

func (f *fakeCertificateClient) PurgeDeletedCertificate(ctx context.Context, name string, options *azcertificates.PurgeDeletedCertificateOptions) (azcertificates.PurgeDeletedCertificateResponse, error) {
	return azcertificates.PurgeDeletedCertificateResponse{}, f.purgeErr
}

func TestGetCertificateNotFound(t *testing.T) {
	client := &fakeCertificateClient{getErr: &azcore.ResponseError{StatusCode: http.StatusNotFound}}
	svc := NewCertificatesServiceWithClient(client)

	_, err := svc.Get(context.Background(), "tls-cert", "")
	if !errors.Is(err, ErrCertificateNotFound) {
		t.Fatalf("expected ErrCertificateNotFound, got %v", err)
	}
}

func TestSetCertificate(t *testing.T) {
	client := &fakeCertificateClient{}
	svc := NewCertificatesServiceWithClient(client)

	err := svc.Set(context.Background(), "tls-cert", CertificateCreateInput{Subject: "CN=tls-cert"})
	if err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	if client.createCall != 1 {
		t.Fatalf("expected one create call, got %d", client.createCall)
	}
}

func TestUpdateCertificate(t *testing.T) {
	client := &fakeCertificateClient{}
	svc := NewCertificatesServiceWithClient(client)
	now := time.Now().UTC().Add(12 * time.Hour)
	enabled := false

	err := svc.Update(context.Background(), "tls-cert", CertificateUpdateInput{Version: "v1", Enabled: &enabled, NotBefore: &now})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}

	if client.updateCall != 1 {
		t.Fatalf("expected one update call, got %d", client.updateCall)
	}
}
