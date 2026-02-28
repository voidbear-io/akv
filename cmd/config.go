package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/frostyeti/akv/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage akv configuration",
		Long:  "Manage akv configuration values. Supports dot (.), colon (:), and forward slash (/) as path separators.",
	}

	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigRmCmd())

	return cmd
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <path>",
		Short: "Get a configuration value",
		Long: `Get a configuration value using path notation.

Path separators: dot (.), colon (:), or forward slash (/)

Examples:
  akv config get vaults.mypvault.url
  akv config get vaults:myvault:url
  akv config get currentVault`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			value := mgr.Get(args[0])
			if value == nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "undefined")
				return nil
			}

			// Output as JSON
			data, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal value: %w", err)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return err
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <path> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value using path notation.

Path separators: dot (.), colon (:), or forward slash (/)

Examples:
  akv config set auth.clientId my-client-id
  akv config set auth:tenantId my-tenant-id
  akv config set currentVault myvault`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.Set(args[0], args[1]); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "set %q = %q\n", args[0], args[1])
			return err
		},
	}
}

func newConfigRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <path>",
		Short: "Remove a configuration value",
		Long: `Remove a configuration value using path notation.

Path separators: dot (.), colon (:), or forward slash (/)

Examples:
  akv config rm vaults.oldvault
  akv config rm auth:servicePrincipal`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.Remove(args[0]); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "removed %q\n", args[0])
			return err
		},
	}
}
