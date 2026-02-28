/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = NewRootCmd()

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// NewRootCmd builds the root command tree for the CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "akv",
		Short:        "Azure Key Vault command-line interface",
		SilenceUsage: true,
		Long: "akv manages Azure Key Vault resources including secrets, keys, " +
			"and certificates.",
	}

	secretsCmd := newSecretsCmd()
	cmd.AddCommand(secretsCmd)
	cmd.AddCommand(newKeysCmd())
	cmd.AddCommand(newCertificatesCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newVaultCmd())
	cmd.AddCommand(newLsCmd()) // Root alias for "secrets ls"

	for _, alias := range newSecretRootAliasCmds() {
		cmd.AddCommand(alias)
	}

	cmd.PersistentFlags().String("vault-url", "", "Azure Key Vault URL (or AKV_VAULT_URL)")

	return cmd
}
