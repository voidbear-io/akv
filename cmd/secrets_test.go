package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/frostyeti/akv/internal/keyvault"
	"github.com/spf13/cobra"
)

type mockSecretService struct {
	getValue string
	getErr   error
	setErr   error
	delErr   error
	updErr   error
	purgeErr error

	setCalls   int
	updCalls   int
	purgeCalls int
	lastName   string
	lastVal    string
	lastVer    string
}

func (m *mockSecretService) Get(ctx context.Context, name string, version string) (string, error) {
	m.lastName = name
	m.lastVer = version
	return m.getValue, m.getErr
}

func (m *mockSecretService) Set(ctx context.Context, name string, value string) error {
	m.setCalls++
	m.lastName = name
	m.lastVal = value
	return m.setErr
}

func (m *mockSecretService) Delete(ctx context.Context, name string) error {
	m.lastName = name
	return m.delErr
}

func (m *mockSecretService) Update(ctx context.Context, name string, in keyvault.SecretUpdateInput) error {
	m.updCalls++
	m.lastName = name
	m.lastVer = in.Version
	return m.updErr
}

func (m *mockSecretService) List(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (m *mockSecretService) Purge(ctx context.Context, name string) error {
	m.purgeCalls++
	m.lastName = name
	return m.purgeErr
}

func TestSecretRootAliasesMatchSecretSubcommands(t *testing.T) {
	tests := []struct {
		name      string
		aliasArgs []string
		subArgs   []string
		setup     func(*mockSecretService)
	}{
		{
			name:      "get alias and subcommand",
			aliasArgs: []string{"get", "db-password", "--version", "a1b2"},
			subArgs:   []string{"secrets", "get", "db-password", "--version", "a1b2"},
			setup: func(s *mockSecretService) {
				s.getValue = "super-secret"
			},
		},
		{
			name:      "set alias and subcommand",
			aliasArgs: []string{"set", "db-password", "new-value"},
			subArgs:   []string{"secrets", "set", "db-password", "new-value"},
		},
		{
			name:      "rm alias and subcommand",
			aliasArgs: []string{"rm", "db-password"},
			subArgs:   []string{"secrets", "rm", "db-password"},
		},
		{
			name:      "ensure alias and subcommand",
			aliasArgs: []string{"ensure", "db-password", "ensured"},
			subArgs:   []string{"secrets", "ensure", "db-password", "ensured"},
			setup: func(s *mockSecretService) {
				s.getErr = keyvault.ErrSecretNotFound
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliasSvc := &mockSecretService{}
			if tt.setup != nil {
				tt.setup(aliasSvc)
			}

			aliasOut, aliasErr := runCommandWithMockService(t, aliasSvc, tt.aliasArgs...)
			if aliasErr != nil {
				t.Fatalf("alias command failed: %v", aliasErr)
			}

			subSvc := &mockSecretService{}
			if tt.setup != nil {
				tt.setup(subSvc)
			}

			subOut, subErr := runCommandWithMockService(t, subSvc, tt.subArgs...)
			if subErr != nil {
				t.Fatalf("subcommand failed: %v", subErr)
			}

			if aliasOut != subOut {
				t.Fatalf("expected alias output %q to match subcommand output %q", aliasOut, subOut)
			}
		})
	}
}

func TestEnsureDoesNotSetWhenSecretExists(t *testing.T) {
	svc := &mockSecretService{getValue: "existing-value"}

	out, err := runCommandWithMockService(t, svc, "ensure", "db-password", "new-value")
	if err != nil {
		t.Fatalf("ensure command failed: %v", err)
	}

	if svc.setCalls != 0 {
		t.Fatalf("expected ensure to skip Set when secret exists")
	}

	if out == "" {
		t.Fatalf("expected ensure output message")
	}
}

func TestGetPassesVersionFlag(t *testing.T) {
	svc := &mockSecretService{getValue: "super-secret"}

	_, err := runCommandWithMockService(t, svc, "secrets", "get", "db-password", "--version", "v-guid")
	if err != nil {
		t.Fatalf("get command failed: %v", err)
	}

	if svc.lastVer != "v-guid" {
		t.Fatalf("expected version to be passed, got %q", svc.lastVer)
	}
}

func TestUpdateCallsService(t *testing.T) {
	svc := &mockSecretService{}

	_, err := runCommandWithMockService(
		t,
		svc,
		"secrets", "update", "db-password",
		"--version", "v-guid",
		"--content-type", "text/plain",
		"--tag", "team=platform",
		"--expires-on", "2030-01-01T00:00:00Z",
		"--set-enabled",
		"--enabled=false",
	)
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	if svc.updCalls != 1 {
		t.Fatalf("expected update call, got %d", svc.updCalls)
	}
	if svc.lastVer != "v-guid" {
		t.Fatalf("expected update version to be set, got %q", svc.lastVer)
	}
}

func TestResolveVaultURLError(t *testing.T) {
	// Clear environment variable to ensure test isolation
	t.Setenv("AKV_VAULT_URL", "")
	// Clear HOME to prevent config file lookup
	t.Setenv("HOME", t.TempDir())

	root := NewRootCmd()
	secretGet, _, err := root.Find([]string{"get"})
	if err != nil {
		t.Fatalf("failed to find get command: %v", err)
	}

	if _, err := resolveVaultURL(secretGet); !errors.Is(err, ErrVaultURLRequired) {
		t.Fatalf("expected ErrVaultURLRequired, got %v", err)
	}
}

func runCommandWithMockService(t *testing.T, service secretService, args ...string) (string, error) {
	t.Helper()

	originalFactory := secretServiceFactory
	secretServiceFactory = func(cmd *cobra.Command) (secretService, error) {
		return service, nil
	}
	t.Cleanup(func() {
		secretServiceFactory = originalFactory
	})

	root := NewRootCmd()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}
