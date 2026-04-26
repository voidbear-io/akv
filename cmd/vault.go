package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/config"
)

func newVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Manage Azure Key Vault configurations",
		Long:  "Manage the list of known Azure Key Vaults and set the default vault to use.",
	}

	cmd.AddCommand(newVaultAddCmd())
	cmd.AddCommand(newVaultRmCmd())
	cmd.AddCommand(newVaultLsCmd())
	cmd.AddCommand(newVaultUseCmd())
	cmd.AddCommand(newVaultShowCmd())

	return cmd
}

func newVaultAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> [url]",
		Short: "Add a new vault",
		Long: `Add a new Azure Key Vault to the configuration.

If URL is not provided, it defaults to https://{name}.vault.azure.net

Examples:
  akv vault add myvault
  akv vault add myvault https://myvault.vault.azure.net`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			name := args[0]
			url := fmt.Sprintf("https://%s.vault.azure.net", name)

			if len(args) > 1 {
				url = args[1]
			}

			// Validate URL format
			if !strings.HasPrefix(url, "https://") || !strings.Contains(url, ".vault.azure.net") {
				return fmt.Errorf("invalid vault URL format: %s", url)
			}

			if err := mgr.AddVault(name, url); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "added vault %q (%s)\n", name, url)
			return err
		},
	}
}

func newVaultRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a vault",
		Long: `Remove a vault from the configuration.

Example:
  akv vault rm myvault`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			name := args[0]
			vault := mgr.GetVault(name)
			if vault == nil {
				return fmt.Errorf("vault %q not found", name)
			}

			if err := mgr.RemoveVault(name); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "removed vault %q\n", name)
			return err
		},
	}
}

func newVaultLsCmd() *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   "ls [pattern]",
		Short: "List vaults",
		Long: `List configured vaults with optional glob pattern filtering.

Examples:
  akv vault ls
  akv vault ls "prod-*"
  akv vault ls "*dev*"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				pattern = args[0]
			}

			vaultNames, err := mgr.ListVaults(pattern)
			if err != nil {
				return err
			}

			currentVault := mgr.GetCurrentVault()

			if len(vaultNames) == 0 {
				if pattern != "" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no vaults match the pattern")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no vaults configured")
				}
				return nil
			}

			for _, name := range vaultNames {
				vault := mgr.GetVault(name)
				prefix := "  "
				if name == currentVault {
					prefix = "* "
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s%s (%s)\n", prefix, name, vault.URL)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&pattern, "filter", "f", "", "Glob pattern to filter vault names")

	return cmd
}

func newVaultUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the current vault",
		Long: `Set the default vault to use for all akv commands.

Example:
  akv vault use myvault`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			name := args[0]
			vault := mgr.GetVault(name)
			if vault == nil {
				return fmt.Errorf("vault %q not found", name)
			}

			if err := mgr.SetCurrentVault(name); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "using vault %q (%s)\n", name, vault.URL)
			return err
		},
	}
}

func newVaultShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [name]",
		Short: "Show vault details",
		Long: `Show details for a vault. If no name is provided, shows the current vault.

Examples:
  akv vault show
  akv vault show myvault`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := config.NewManager()
			if err != nil {
				return err
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			} else {
				name = mgr.GetCurrentVault()
				if name == "" {
					return fmt.Errorf("no vault specified and no current vault set")
				}
			}

			vault := mgr.GetVault(name)
			if vault == nil {
				return fmt.Errorf("vault %q not found", name)
			}

			current := ""
			if mgr.GetCurrentVault() == name {
				current = " (current)"
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Name: %s%s\nURL:  %s\n", name, current, vault.URL)
			return err
		},
	}
}
