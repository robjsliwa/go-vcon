package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

// ====================== root initialisation =========================

var rootCmd = &cobra.Command{
	Use:   "vconctl",
	Short: "vconctl - a tool for working with vCon files",
	Long:  `vconctl is a command-line utility for validating, signing, encrypting, verifying, and decrypting vCon (Virtual Conversation) files.`,
}

func main() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(validateCmd, signCmd, encryptCmd, verifyCmd, decryptCmd, genkeyCmd)

	// flags
	signCmd.Flags().StringP("key", "k", "", "Path to private key file (required)")
	signCmd.Flags().StringP("cert", "c", "", "Path to certificate file (required)")
	signCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to <file>.signed.json)")

	encryptCmd.Flags().StringP("cert", "c", "", "Path to recipient certificate (required)")
	encryptCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to <file>.encrypted.json)")

	verifyCmd.Flags().StringP("cert", "c", "", "Path to trust anchor (leaf or CA) (required)")

	decryptCmd.Flags().StringP("key", "k", "", "Path to private key file (required)")
	decryptCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to <file>.decrypted.json)")

	genkeyCmd.Flags().StringP("key", "k", "", "Output private-key path (default: test_key.pem)")
	genkeyCmd.Flags().StringP("cert", "c", "", "Output certificate path (default: test_cert.pem)")
}

// ============================ validate ==============================

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a vCon file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		for _, p := range args {
			fmt.Printf("Validating %s…\n", p)
			if _, err := vcon.LoadFromFile(p, vcon.PropertyHandlingStrict); err != nil {
				fmt.Printf("❌ %v\n", err)
				continue
			}
			fmt.Printf("✅ %s is valid\n", p)
		}
	},
}

// ============================== sign ================================

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

	// ---------- read & unmarshal vCon ----------
	raw, err := os.ReadFile(path)
	if err != nil {
		die("reading vCon", err)
	}
	var v vcon.VCon
	if err := json.Unmarshal(raw, &v); err != nil {
		die("parsing JSON", err)
	}

	// ---------- keys & cert ----------
	priv := readPrivateKey(keyPath)
	cert := readCertificate(certPath)

	// ---------- sign ----------
	signed, err := v.Sign(priv, []*x509.Certificate{cert})
	if err != nil {
		die("signing vCon", err)
	}

	// ---------- output ----------
	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".signed" + ext
	}
	if err := writeJSON(outPath, signed.JSON); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Signed vCon written to %s\n", outPath)
}

// ============================= verify ===============================

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

	// -------- load bare JWS --------
	jwsMap := readBareJWS(path)

	// -------- trust anchors --------
	root := x509.NewCertPool()
	if ok := appendPEMToPool(root, caPath); !ok {
		die("loading trust anchor", fmt.Errorf("invalid PEM in %s", caPath))
	}

	// -------- verify --------
	signed := vcon.SignedVCon{JSON: jwsMap}
	vc, err := signed.Verify(root)
	if err != nil {
		die("signature verification failed", err)
	}

	fmt.Println("✅ Signature verified!")
	fmt.Printf("Subject : %s\nUUID    : %s\nCreated : %s\nParties : %d\n",
		vc.Subject, vc.UUID, vc.CreatedAt, len(vc.Parties))
}

// ============================ encrypt ===============================

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

// ========================= decrypt & misc ===========================

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

	// Extract the JWE content from the wrapper
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

	// Decrypt
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

// ============================== genkey ===============================

var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate a test RSA key pair and self-signed certificate",
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")
		certPath, _ := cmd.Flags().GetString("cert")
		if keyPath == "" {
			keyPath = "test_key.pem"
		}
		if certPath == "" {
			certPath = "test_cert.pem"
		}
		generateKeyPair(keyPath, certPath)
	},
}

func generateKeyPair(keyPath, certPath string) {
	fmt.Printf("Generating RSA key pair and certificate…\n")

	// Generate RSA private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		die("generating private key", err)
	}

	// Create certificate template
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour) // Valid for 1 year
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		die("generating serial number", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"Test Organization"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "test.example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		die("creating certificate", err)
	}

	// Encode private key to PKCS#8 PEM format
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		die("marshaling private key", err)
	}
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	// Encode certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Write private key to file
	if err := os.WriteFile(keyPath, privKeyPEM, 0600); err != nil {
		die("writing private key", err)
	}

	// Write certificate to file
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		die("writing certificate", err)
	}

	fmt.Printf("✅ Private key written to %s\n", keyPath)
	fmt.Printf("✅ Certificate written to %s\n", certPath)
}

// ========================= helper utils ============================

func readBareJWS(path string) map[string]any {
	raw, err := os.ReadFile(path)
	if err != nil {
		die("reading file", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		die("parsing JSON", err)
	}
	return m
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func readPrivateKey(p string) *rsa.PrivateKey {
	raw, err := os.ReadFile(p)
	if err != nil {
		die("reading private key", err)
	}
	b, _ := pem.Decode(raw)
	if b == nil {
		die("decoding PEM", fmt.Errorf("no block found"))
	}

	switch b.Type {
	case "RSA PRIVATE KEY":
		k, err := x509.ParsePKCS1PrivateKey(b.Bytes)
		if err != nil {
			die("PKCS1 parse", err)
		}
		return k
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(b.Bytes)
		if err != nil {
			die("PKCS8 parse", err)
		}
		if rsaK, ok := k.(*rsa.PrivateKey); ok {
			return rsaK
		}
	}
	die("private key", fmt.Errorf("unsupported key type %q", b.Type))
	return nil
}

func readCertificate(p string) *x509.Certificate {
	raw, err := os.ReadFile(p)
	if err != nil {
		die("reading certificate", err)
	}
	b, _ := pem.Decode(raw)
	if b == nil || b.Type != "CERTIFICATE" {
		die("certificate", fmt.Errorf("invalid PEM"))
	}
	c, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		die("parsing certificate", err)
	}
	return c
}

func appendPEMToPool(pool *x509.CertPool, pemPath string) bool {
	raw, err := os.ReadFile(pemPath)
	if err != nil {
		die("reading CA file", err)
	}
	return pool.AppendCertsFromPEM(raw)
}

func die(context string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s: %v\n", context, err)
	os.Exit(1)
}
