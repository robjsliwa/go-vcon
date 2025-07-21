package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

// Command: sign

var signCmd = &cobra.Command{
	Use:   "sign [file]",
	Short: "Sign a vCon file using a private key and certificate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")
		certPath, _ := cmd.Flags().GetString("cert")
		outPath, _ := cmd.Flags().GetString("output")
		if keyPath == "" || certPath == "" {
			fmt.Println("Error: --key and --cert are required")
			_ = cmd.Help()
			os.Exit(1)
		}
		signFile(args[0], keyPath, certPath, outPath)
	},
}

func signFile(path, keyPath, certPath, outPath string) {
	fmt.Printf("Signing %s…\n", path)

	raw, err := os.ReadFile(path)
	if err != nil {
		die("reading vCon", err)
	}
	var v vcon.VCon
	if err := json.Unmarshal(raw, &v); err != nil {
		die("parsing JSON", err)
	}

	priv := readPrivateKey(keyPath)
	cert := readCertificate(certPath)

	signed, err := v.Sign(priv, []*x509.Certificate{cert})
	if err != nil {
		die("signing vCon", err)
	}

	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".signed" + ext
	}
	if err := writeJSON(outPath, signed.JSON); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Signed vCon written to %s\n", outPath)
}

// Command: verify

var verifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verify the signature on a signed vCon",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		caPath, _ := cmd.Flags().GetString("cert")
		if caPath == "" {
			fmt.Println("Error: --cert is required")
			_ = cmd.Help()
			os.Exit(1)
		}
		verifyFile(args[0], caPath)
	},
}

func verifyFile(path, caPath string) {
	fmt.Printf("Verifying %s…\n", path)

	jwsMap := readBareJWS(path)

	root := x509.NewCertPool()
	if ok := appendPEMToPool(root, caPath); !ok {
		die("loading trust anchor", fmt.Errorf("invalid PEM in %s", caPath))
	}

	signed := vcon.SignedVCon{JSON: jwsMap}
	vc, err := signed.Verify(root)
	if err != nil {
		die("signature verification failed", err)
	}

	fmt.Println("✅ Signature verified!")
	fmt.Printf("Subject : %s\nUUID    : %s\nCreated : %s\nParties : %d\n",
		vc.Subject, vc.UUID, vc.CreatedAt, len(vc.Parties))
}