package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
)

func newCertificateUploadCmd() *cobra.Command {
	var (
		name    string
		passwd  string
		tagVals []string
	)

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
			if err != nil {
				return err
			}

			input := args[0]
			if name == "" {
				name = strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
			}

			data, err := os.ReadFile(input)
			if err != nil {
				return err
			}

			tagMap, err := parseTags(tagVals)
			if err != nil {
				return err
			}

			payload := base64.StdEncoding.EncodeToString(data)
			if err := service.ImportCertificate(cmd.Context(), name, keyvault.CertificateImportInput{Base64EncodedCertificate: payload, Password: passwd, Tags: tagMap}); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "uploaded certificate %q\n", name)
			return err
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Certificate name; defaults to input file name")
	cmd.Flags().StringVar(&passwd, "password", "", "Password for encrypted PFX input")
	cmd.Flags().StringArrayVar(&tagVals, "tag", nil, "Certificate tag in key=value format")
	return cmd
}
