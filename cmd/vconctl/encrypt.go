package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-jose/go-jose/v4"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

// Command: encrypt

var encryptCmd = &cobra.Command{
	Use:   "encrypt [file]",
	Short: "Encrypt a signed vCon for one recipient",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certPath, _ := cmd.Flags().GetString("cert")
		outPath, _ := cmd.Flags().GetString("output")
		if certPath == "" {
			fmt.Println("Error: --cert is required")
			_ = cmd.Help()
			os.Exit(1)
		}
		encryptFile(args[0], certPath, outPath)
	},
}

func encryptFile(path, certPath, outPath string) {
	fmt.Printf("Encrypting %s…\n", path)

	jwsMap := readBareJWS(path)
	signed := vcon.SignedVCon{JSON: jwsMap}
	cert := readCertificate(certPath)

	obj, err := signed.Encrypt([]jose.Recipient{{
		Algorithm: jose.RSA_OAEP,
		Key:       cert.PublicKey,
	}})
	if err != nil {
		die("encrypting", err)
	}

	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".encrypted" + ext
	}
	if err := writeJSON(outPath, obj); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Encrypted vCon written to %s\n", outPath)
}

// Command decrypt

var decryptCmd = &cobra.Command{
	Use:   "decrypt [file]",
	Short: "Decrypt an encrypted vCon file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")
		outPath, _ := cmd.Flags().GetString("output")
		if keyPath == "" {
			fmt.Println("Error: --key is required")
			_ = cmd.Help()
			os.Exit(1)
		}
		decryptFile(args[0], keyPath, outPath)
	},
}

func decryptFile(path, keyPath, outPath string) {
	fmt.Printf("Decrypting %s…\n", path)

	// Read encrypted JWE
	raw, err := os.ReadFile(path)
	if err != nil {
		die("reading file", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		die("parsing JSON", err)
	}

	jweContent, ok := m["jwe"]
	if !ok {
		die("extracting JWE", fmt.Errorf("no 'jwe' field found"))
	}
	jweMap, ok := jweContent.(map[string]any)
	if !ok {
		die("extracting JWE", fmt.Errorf("'jwe' field is not an object"))
	}

	encrypted := vcon.EncryptedVCon{JSON: jweMap}
	priv := readPrivateKey(keyPath)

	decrypted, err := encrypted.Decrypt(priv)
	if err != nil {
		die("decrypting", err)
	}

	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".decrypted" + ext
	}
	if err := writeJSON(outPath, decrypted); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Decrypted vCon written to %s\n", outPath)
}