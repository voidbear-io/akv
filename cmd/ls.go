package cmd

import (
	"fmt"

	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// newLsCmd creates the root 'ls' alias command that maps to 'secrets ls'.
func newLsCmd() *cobra.Command {
	var pattern string

	cmd := &cobra.Command{
		Use:   "ls [pattern]",
		Short: "List secrets (alias for 'akv secrets ls')",
		Long: `List secrets in the Azure Key Vault.

This is a root-level alias for 'akv secrets ls'.

Examples:
  akv ls
  akv ls "my-secret-*"
  akv ls "*prod*"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				pattern = args[0]
			}

			service, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			secrets, err := service.List(cmd.Context())
			if err != nil {
				return err
			}

			// Apply pattern filter if provided
			var filtered []string
			if pattern != "" {
				g, err := glob.Compile(pattern)
				if err != nil {
					return fmt.Errorf("invalid pattern: %w", err)
				}
				for _, secret := range secrets {
					if g.Match(secret) {
						filtered = append(filtered, secret)
					}
				}
			} else {
				filtered = secrets
			}

			if len(filtered) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No secrets found.")
				return nil
			}

			for _, secret := range filtered {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), secret)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&pattern, "filter", "f", "", "Glob pattern to filter secrets")

	return cmd
}
