package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/go/dotenv"
)

func newSecretGetCmd() *cobra.Command {
	var version string
	var format string

	cmd := &cobra.Command{
		Use:   "get <name>...",
		Short: "Get a secret value",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			values := make(map[string]string, len(args))
			for _, name := range args {
				value, err := service.Get(cmd.Context(), name, version)
				if err != nil {
					return err
				}
				values[name] = value
			}

			return writeSecretOutput(cmd, values, format)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific secret version GUID")
	cmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json, sh, bash, zsh, powershell, pwsh, dotenv, azure-devops, github-actions, cast, null)")
	return cmd
}

func newSecretGetAliasCmd() *cobra.Command {
	return newSecretGetCmd()
}

func writeSecretOutput(cmd *cobra.Command, values map[string]string, format string) error {
	if format == "" {
		format = "text"
	}

	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal secrets to JSON: %w", err)
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	case "null-terminated", "null":
		for _, value := range values {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\x00", value); err != nil {
				return err
			}
		}
		return nil
	case "sh", "bash", "zsh":
		for key, value := range values {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "export %s='%s'\n", toScreamingSnakeCase(key), value); err != nil {
				return err
			}
		}
		return nil
	case "powershell", "pwsh":
		for key, value := range values {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "$Env:%s='%s'\n", toScreamingSnakeCase(key), value); err != nil {
				return err
			}
		}
		return nil
	case "dotenv", "env", ".env":
		doc := dotenv.NewDoc()
		for key, value := range values {
			doc.Set(toScreamingSnakeCase(key), value)
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), doc.String())
		return err
	case "azure-pipelines", "ado", "azure-devops":
		for key, value := range values {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "##vso[task.setvariable variable=%s;]%s\n", toScreamingSnakeCase(key), value); err != nil {
				return err
			}
		}
		return nil
	case "cast", "castfile":
		envFile := os.Getenv("CAST_SECRETS")
		if envFile == "" {
			return fmt.Errorf("CAST_SECRETS environment variable is not set")
		}
		if err := os.MkdirAll(filepath.Dir(envFile), 0700); err != nil {
			return fmt.Errorf("creating directory for CAST_SECRETS file failed: %w", err)
		}
		f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("opening CAST_SECRETS file failed: %w", err)
		}
		defer func() { _ = f.Close() }()
		for key, value := range values {
			if strings.ContainsAny(value, "\r\n") {
				if _, err := fmt.Fprintf(f, "%s<< EOF\n%s\nEOF\n", toScreamingSnakeCase(key), value); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(f, "%s=%s\n", toScreamingSnakeCase(key), value); err != nil {
				return err
			}
		}
		return nil
	default:
		for _, value := range values {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), value); err != nil {
				return err
			}
		}
		return nil
	}
}
