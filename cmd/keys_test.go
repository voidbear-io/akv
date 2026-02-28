package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/frostyeti/akv/internal/keyvault"
	"github.com/spf13/cobra"
)

type mockKeyService struct {
	getInfo  keyvault.KeyInfo
	getErr   error
	setErr   error
	updErr   error
	delErr   error
	purgeErr error

	lastVersion string
	updateCalls int
}

func (m *mockKeyService) Get(ctx context.Context, name string, version string) (keyvault.KeyInfo, error) {
	m.lastVersion = version
	return m.getInfo, m.getErr
}

func (m *mockKeyService) Set(ctx context.Context, name string, in keyvault.KeyCreateInput) error {
	return m.setErr
}

func (m *mockKeyService) Update(ctx context.Context, name string, in keyvault.KeyUpdateInput) error {
	m.updateCalls++
	m.lastVersion = in.Version
	return m.updErr
}

func (m *mockKeyService) Delete(ctx context.Context, name string) error {
	return m.delErr
}

func (m *mockKeyService) Purge(ctx context.Context, name string) error {
	return m.purgeErr
}

func TestKeysGetPassesVersion(t *testing.T) {
	svc := &mockKeyService{getInfo: keyvault.KeyInfo{ID: "kid", Type: "RSA"}}

	original := keyServiceFactory
	keyServiceFactory = func(cmd *cobra.Command) (keyService, error) { return svc, nil }
	t.Cleanup(func() { keyServiceFactory = original })

	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"keys", "get", "app-key", "--version", "v1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("keys get failed: %v", err)
	}

	if svc.lastVersion != "v1" {
		t.Fatalf("expected version v1, got %q", svc.lastVersion)
	}
}

func TestKeysUpdatePassesVersion(t *testing.T) {
	svc := &mockKeyService{}

	original := keyServiceFactory
	keyServiceFactory = func(cmd *cobra.Command) (keyService, error) { return svc, nil }
	t.Cleanup(func() { keyServiceFactory = original })

	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"keys", "update", "app-key", "--version", "v3", "--tag", "team=platform"})

	if err := root.Execute(); err != nil {
		t.Fatalf("keys update failed: %v", err)
	}

	if svc.updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", svc.updateCalls)
	}
	if svc.lastVersion != "v3" {
		t.Fatalf("expected version v3, got %q", svc.lastVersion)
	}
}
