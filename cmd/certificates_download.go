package cmd

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidbear-io/akv/internal/keyvault"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

func newCertificateDownloadCmd() *cobra.Command {
	var (
		version string
		format  string
		output  string
		passwd  string
	)

	cmd := &cobra.Command{
		Use:   "download <name>",
		Short: "Download a certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := certificateServiceFactory(cmd)
			if err != nil {
				return err
			}

			secretSvc, err := secretServiceFactory(cmd)
			if err != nil {
				return err
			}

			info, err := service.Get(cmd.Context(), args[0], version)
			if err != nil {
				return err
			}

			secret, err := secretSvc.Get(cmd.Context(), args[0], version)
			if err != nil {
				return err
			}

			data, err := renderCertificateDownload(info, secret, format, passwd)
			if err != nil {
				return err
			}

			if output != "" {
				return os.WriteFile(output, data, 0600)
			}

			_, err = cmd.OutOrStdout().Write(data)
			if err == nil && len(data) > 0 && data[len(data)-1] != '\n' {
				_, err = io.WriteString(cmd.OutOrStdout(), "\n")
			}
			return err
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Specific certificate version")
	cmd.Flags().StringVar(&format, "format", "pem", "Output format: pem, pfx, cer")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&passwd, "password", "", "Password to use when writing a PFX file")
	return cmd
}

func renderCertificateDownload(info keyvault.CertificateInfo, secret string, format string, password string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("decode certificate secret: %w", err)
	}

	switch strings.ToLower(format) {
	case "cer":
		if block, _ := pem.Decode(data); block != nil {
			data = block.Bytes
		}
		return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data}), nil
	case "pem":
		if block, _ := pem.Decode(data); block != nil {
			return data, nil
		}
		return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data}), nil
	case "pfx", "p12":
		if password == "" {
			password = pkcs12.DefaultPassword
		}
		cert, key, chain, err := decodeCertificateAndKey(data, password)
		if err != nil {
			return nil, err
		}
		return pkcs12.Modern2023.Encode(key, cert, chain, password)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func decodeCertificateAndKey(data []byte, password string) (*x509.Certificate, any, []*x509.Certificate, error) {
	key, cert, chain, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode pfx: %w", err)
	}
	if cert == nil {
		return nil, nil, nil, fmt.Errorf("pfx has no certificate")
	}
	return cert, key, chain, nil
}
