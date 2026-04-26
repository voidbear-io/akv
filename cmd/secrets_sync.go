package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

type secretSync struct {
	Value     *string           `json:"value"`
	Ensure    bool              `json:"ensure"`
	Size      int               `json:"size"`
	NoUpper   bool              `json:"noUpper"`
	NoLower   bool              `json:"noLower"`
	NoDigits  bool              `json:"noDigits"`
	NoSpecial bool              `json:"noSpecial"`
	Special   string            `json:"special"`
	Chars     string            `json:"chars"`
	Delete    bool              `json:"delete"`
	Tags      map[string]string `json:"tags"`
}

func newSecretsSyncCmd() *cobra.Command {
	var (
		file   string
		stdin  bool
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync secrets from JSON",
		Long: `Sync secrets from JSON.

Input shape:
  {
    "db-password": {
      "value": "new-value",
      "delete": false,
      "tags": {"owner": "app"}
    },
    "legacy-secret": {
      "delete": true
    }
  }`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" && !stdin {
				return fmt.Errorf("must specify either --file or --stdin")
			}
			if file != "" && stdin {
				return fmt.Errorf("--file and --stdin are mutually exclusive")
			}

			payload, err := readJSONInput(file, stdin)
			if err != nil {
				return err
			}

			var raw map[string]json.RawMessage
			if err := json.Unmarshal(payload, &raw); err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}

			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			for name, data := range raw {
				if err := syncSecret(cmd, service, name, data, dryRun); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Input JSON file path (JSON format)")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read JSON from stdin (JSON format)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making changes")
	return cmd
}

func syncSecret(cmd *cobra.Command, service secretService, name string, data []byte, dryRun bool) error {
	var simple string
	if err := json.Unmarshal(data, &simple); err == nil {
		return syncSecretValue(cmd, service, name, &simple, nil, dryRun)
	}

	var obj secretSync
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("parse secret %q: %w", name, err)
	}

	return syncSecretValue(cmd, service, name, obj.Value, &obj, dryRun)
}

func syncSecretValue(cmd *cobra.Command, service secretService, name string, value *string, obj *secretSync, dryRun bool) error {
	current, err := service.GetData(cmd.Context(), name, "")
	if err != nil && !errorsIsSecretNotFound(err) {
		return err
	}

	if err == nil && value != nil && current.Value != nil && *current.Value == *value {
		return nil
	}

	if dryRun {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "would sync secret %q\n", name)
		return err
	}

	if value != nil {
		if err := service.Set(cmd.Context(), name, *value); err != nil {
			return err
		}
	}

	if obj != nil && obj.Delete {
		if err := service.Delete(cmd.Context(), name); err != nil && !errorsIsSecretNotFound(err) {
			return err
		}
	}

	return nil
}

func errorsIsSecretNotFound(err error) bool {
	return errors.Is(err, keyvault.ErrSecretNotFound)
}
