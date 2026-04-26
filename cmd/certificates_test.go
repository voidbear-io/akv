package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

type mockCertificateService struct {
	getInfo  keyvault.CertificateInfo
	getErr   error
	setErr   error
	listVals []string
	updErr   error
	delErr   error
	purgeErr error

	lastVersion string
	updateCalls int
}

func (m *mockCertificateService) Get(ctx context.Context, name string, version string) (keyvault.CertificateInfo, error) {
	m.lastVersion = version
	return m.getInfo, m.getErr
}

func (m *mockCertificateService) Set(ctx context.Context, name string, in keyvault.CertificateCreateInput) error {
	return m.setErr
}

func (m *mockCertificateService) ImportCertificate(ctx context.Context, name string, in keyvault.CertificateImportInput) error {
	return nil
}

func (m *mockCertificateService) List(ctx context.Context) ([]string, error) {
	return m.listVals, nil
}

func (m *mockCertificateService) Update(ctx context.Context, name string, in keyvault.CertificateUpdateInput) error {
	m.updateCalls++
	m.lastVersion = in.Version
	return m.updErr
}

func (m *mockCertificateService) Delete(ctx context.Context, name string) error {
	return m.delErr
}

func (m *mockCertificateService) Purge(ctx context.Context, name string) error {
	return m.purgeErr
}

func TestCertificatesGetPassesVersion(t *testing.T) {
	svc := &mockCertificateService{getInfo: keyvault.CertificateInfo{ID: "cid", ContentType: "application/x-pem-file"}}

	original := certificateServiceFactory
	certificateServiceFactory = func(cmd *cobra.Command) (certificateService, error) { return svc, nil }
	t.Cleanup(func() { certificateServiceFactory = original })

	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"certificates", "get", "tls-cert", "--version", "v2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("certificates get failed: %v", err)
	}

	if svc.lastVersion != "v2" {
		t.Fatalf("expected version v2, got %q", svc.lastVersion)
	}
}

func TestCertificatesUpdatePassesVersion(t *testing.T) {
	svc := &mockCertificateService{}

	original := certificateServiceFactory
	certificateServiceFactory = func(cmd *cobra.Command) (certificateService, error) { return svc, nil }
	t.Cleanup(func() { certificateServiceFactory = original })

	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"certificates", "update", "tls-cert", "--version", "v4", "--tag", "team=security"})

	if err := root.Execute(); err != nil {
		t.Fatalf("certificates update failed: %v", err)
	}

	if svc.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", svc.updateCalls)
	}
	if svc.lastVersion != "v4" {
		t.Fatalf("expected version v4, got %q", svc.lastVersion)
	}
}

func TestCertificatesListFiltersPatterns(t *testing.T) {
	svc := &mockCertificateService{listVals: []string{"tls-prod", "tls-dev", "ca-prod"}}

	original := certificateServiceFactory
	certificateServiceFactory = func(cmd *cobra.Command) (certificateService, error) { return svc, nil }
	t.Cleanup(func() { certificateServiceFactory = original })

	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"certificates", "ls", "tls-*"})

	if err := root.Execute(); err != nil {
		t.Fatalf("certificates ls failed: %v", err)
	}
}
