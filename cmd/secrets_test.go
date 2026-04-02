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
	listVals []string

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

func (m *mockSecretService) GetData(ctx context.Context, name string, version string) (keyvault.SecretInfo, error) {
	m.lastName = name
	m.lastVer = version
	if m.getErr != nil {
		return keyvault.SecretInfo{}, m.getErr
	}
	return keyvault.SecretInfo{}, nil
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
	return m.listVals, nil
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
			aliasArgs: []string{"ensure", "db-password"},
			subArgs:   []string{"secrets", "ensure", "db-password"},
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

	out, err := runCommandWithMockService(t, svc, "ensure", "db-password")
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

func TestEnsureGeneratesWhenMissing(t *testing.T) {
	svc := &mockSecretService{getErr: keyvault.ErrSecretNotFound}

	out, err := runCommandWithMockService(t, svc, "ensure", "db-password", "--size", "8", "--chars", "abc")
	if err != nil {
		t.Fatalf("ensure command failed: %v", err)
	}

	if svc.setCalls != 1 {
		t.Fatalf("expected ensure to set missing secret")
	}

	if out == "" {
		t.Fatalf("expected generated output message")
	}
}

func TestLsListsAndFiltersSecrets(t *testing.T) {
	svc := &mockSecretService{listVals: []string{"app-prod", "app-dev", "db-prod"}}

	out, err := runCommandWithMockService(t, svc, "ls", "app-*")
	if err != nil {
		t.Fatalf("ls command failed: %v", err)
	}

	if out == "" {
		t.Fatalf("expected ls output")
	}

	if svc.setCalls != 0 {
		t.Fatalf("unexpected set call during ls")
	}
}

func TestGenerateSecretValueUsesCustomChars(t *testing.T) {
	secret, err := generateSecretValue(secretGenerationOptions{Size: 12, Chars: "ab"})
	if err != nil {
		t.Fatalf("generateSecretValue returned error: %v", err)
	}

	if len(secret) != 12 {
		t.Fatalf("expected generated secret length 12, got %d", len(secret))
	}

	for _, r := range secret {
		if r != 'a' && r != 'b' {
			t.Fatalf("unexpected character %q in generated secret", r)
		}
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

func TestGetSupportsJsonOutput(t *testing.T) {
	svc := &mockSecretService{getValue: "super-secret"}

	out, err := runCommandWithMockService(t, svc, "secrets", "get", "db-password", "--format", "json")
	if err != nil {
		t.Fatalf("get command failed: %v", err)
	}

	if out == "" {
		t.Fatalf("expected json output")
	}
}

func TestGetSupportsShellOutput(t *testing.T) {
	svc := &mockSecretService{getValue: "super-secret"}

	out, err := runCommandWithMockService(t, svc, "secrets", "get", "db-password", "--format", "bash")
	if err != nil {
		t.Fatalf("get command failed: %v", err)
	}

	if out == "" {
		t.Fatalf("expected shell output")
	}
}

func TestGetDataCommandExists(t *testing.T) {
	svc := &mockSecretService{}

	_, err := runCommandWithMockService(t, svc, "secrets", "get-data", "db-password")
	if err != nil {
		t.Fatalf("get-data command failed: %v", err)
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
	t.Setenv("AKV_VAULT", "")
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

func TestResolveVaultURLExpandsShortName(t *testing.T) {
	t.Setenv("AKV_VAULT", "my-vault")
	t.Setenv("AKV_VAULT_URL", "")
	t.Setenv("HOME", t.TempDir())

	root := NewRootCmd()
	secretGet, _, err := root.Find([]string{"get"})
	if err != nil {
		t.Fatalf("failed to find get command: %v", err)
	}

	url, err := resolveVaultURL(secretGet)
	if err != nil {
		t.Fatalf("resolveVaultURL returned error: %v", err)
	}

	if url != "https://my-vault.vault.azure.net" {
		t.Fatalf("expected expanded vault URL, got %q", url)
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
