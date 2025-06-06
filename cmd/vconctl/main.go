package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-jose/go-jose/v4"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vconctl",
	Short: "vconctl - a tool for working with vCon files",
	Long: `vconctl is a command line utility for validating, signing, encrypting,
verifying, and decrypting vCon (Virtual Conversation) files.`,
}

func main() {
    rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add commands to the root command
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(decryptCmd)
}

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a vCon file",
	Long:  `Validate checks if a vCon file conforms to the vCon specification.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, path := range args {
			validateFile(path)
		}
	},
}

// signCmd represents the sign command
var signCmd = &cobra.Command{
	Use:   "sign [file]",
	Short: "Sign a vCon file",
	Long:  `Sign a vCon file using a private key and certificate.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		privateKeyPath, _ := cmd.Flags().GetString("key")
		certPath, _ := cmd.Flags().GetString("cert")
		outputPath, _ := cmd.Flags().GetString("output")

		if privateKeyPath == "" || certPath == "" {
			fmt.Println("Error: Private key and certificate paths are required")
			cmd.Help()
			os.Exit(1)
		}

		signFile(args[0], privateKeyPath, certPath, outputPath)
	},
}

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt [file]",
	Short: "Encrypt a vCon file",
	Long:  `Encrypt a vCon file for secure storage and transmission.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certPath, _ := cmd.Flags().GetString("cert")
		outputPath, _ := cmd.Flags().GetString("output")

		if certPath == "" {
			fmt.Println("Error: Certificate path is required")
			cmd.Help()
			os.Exit(1)
		}

		encryptFile(args[0], certPath, outputPath)
	},
}

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verify a signed vCon file",
	Long:  `Verify the signature of a signed vCon file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certPath, _ := cmd.Flags().GetString("cert")

		if certPath == "" {
			fmt.Println("Error: Certificate path is required")
			cmd.Help()
			os.Exit(1)
		}

		verifyFile(args[0], certPath)
	},
}

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt [file]",
	Short: "Decrypt an encrypted vCon file",
	Long:  `Decrypt an encrypted vCon file using a private key.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		privateKeyPath, _ := cmd.Flags().GetString("key")
		outputPath, _ := cmd.Flags().GetString("output")

		if privateKeyPath == "" {
			fmt.Println("Error: Private key path is required")
			cmd.Help()
			os.Exit(1)
		}

		decryptFile(args[0], privateKeyPath, outputPath)
	},
}

func init() {
	// Add flags to sign command
	signCmd.Flags().StringP("key", "k", "", "Path to private key file (required)")
	signCmd.Flags().StringP("cert", "c", "", "Path to certificate file (required)")
	signCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to input file with .signed.json extension)")

	// Add flags to encrypt command
	encryptCmd.Flags().StringP("cert", "c", "", "Path to certificate file (required)")
	encryptCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to input file with .encrypted.json extension)")

	// Add flags to verify command
	verifyCmd.Flags().StringP("cert", "c", "", "Path to certificate or CA file (required)")

	// Add flags to decrypt command
	decryptCmd.Flags().StringP("key", "k", "", "Path to private key file (required)")
	decryptCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to input file with .decrypted.json extension)")
}

// validateFile validates a vCon file
func validateFile(path string) {
	fmt.Printf("Validating %s...\n", path)

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		return
	}

	var v vcon.VCon
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Printf("❌ Error parsing JSON: %v\n", err)
		return
	}

	if err := v.Validate(); err != nil {
		fmt.Printf("❌ Invalid vCon: %v\n", err)
		return
	}

	fmt.Printf("✅ %s is a valid vCon file\n", path)
}

// signFile signs a vCon file
func signFile(path, keyPath, certPath, outputPath string) {
	fmt.Printf("Signing %s...\n", path)

	// Read the vCon file
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		return
	}

	var v vcon.VCon
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Printf("❌ Error parsing JSON: %v\n", err)
		return
	}

	// Read the private key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		fmt.Printf("❌ Error reading private key: %v\n", err)
		return
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY" {
		fmt.Println("❌ Failed to decode PEM block containing private key")
		return
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		var parsedKey interface{}
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			fmt.Printf("❌ Error parsing private key: %v\n", err)
			return
		}

		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			fmt.Println("❌ Private key is not an RSA key")
			return
		}
	}

	// Read the certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Printf("❌ Error reading certificate: %v\n", err)
		return
	}

	block, _ = pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Println("❌ Failed to decode PEM block containing certificate")
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("❌ Error parsing certificate: %v\n", err)
		return
	}

	// Sign the vCon
	signed, err := v.Sign(privateKey, []*x509.Certificate{cert})
	if err != nil {
		fmt.Printf("❌ Error signing vCon: %v\n", err)
		return
	}

	// Marshal the signed vCon
	signedData, err := json.MarshalIndent(signed, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling signed vCon: %v\n", err)
		return
	}

	// Determine output path
	if outputPath == "" {
		ext := filepath.Ext(path)
		outputPath = path[:len(path)-len(ext)] + ".signed" + ext
	}

	// Write the signed vCon
	if err := os.WriteFile(outputPath, signedData, 0644); err != nil {
		fmt.Printf("❌ Error writing signed vCon: %v\n", err)
		return
	}

	fmt.Printf("✅ Signed vCon written to %s\n", outputPath)
}

// encryptFile encrypts a vCon file
func encryptFile(path, certPath, outputPath string) {
	fmt.Printf("Encrypting %s...\n", path)

	// Read the vCon file
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		return
	}

	var signedVCon vcon.SignedVCon
	if err := json.Unmarshal(data, &signedVCon); err != nil {
		fmt.Printf("❌ Error parsing signed vCon: %v\n", err)
		return
	}

	// Read the certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Printf("❌ Error reading certificate: %v\n", err)
		return
	}

	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Println("❌ Failed to decode PEM block containing certificate")
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		fmt.Printf("❌ Error parsing certificate: %v\n", err)
		return
	}

	// Create recipient for encryption
	recipient := jose.Recipient{
		Algorithm: jose.RSA_OAEP,
		Key:       cert.PublicKey,
	}

	// Encrypt the vCon
	encrypted, err := signedVCon.Encrypt([]jose.Recipient{recipient})
	if err != nil {
		fmt.Printf("❌ Error encrypting vCon: %v\n", err)
		return
	}

	// Marshal the encrypted vCon
	encryptedData, err := json.MarshalIndent(encrypted, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling encrypted vCon: %v\n", err)
		return
	}

	// Determine output path
	if outputPath == "" {
		ext := filepath.Ext(path)
		outputPath = path[:len(path)-len(ext)] + ".encrypted" + ext
	}

	// Write the encrypted vCon
	if err := os.WriteFile(outputPath, encryptedData, 0644); err != nil {
		fmt.Printf("❌ Error writing encrypted vCon: %v\n", err)
		return
	}

	fmt.Printf("✅ Encrypted vCon written to %s\n", outputPath)
}

// verifyFile verifies a signed vCon file
func verifyFile(path, certPath string) {
	fmt.Printf("Verifying %s...\n", path)

	// Read the vCon file
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		return
	}

	var signedVCon vcon.SignedVCon
	if err := json.Unmarshal(data, &signedVCon); err != nil {
		fmt.Printf("❌ Error parsing signed vCon: %v\n", err)
		return
	}

	// Read the certificate
	certData, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Printf("❌ Error reading certificate: %v\n", err)
		return
	}

	// Create a root pool
	rootPool := x509.NewCertPool()

	if ok := rootPool.AppendCertsFromPEM(certData); !ok {
		fmt.Println("❌ Failed to add certificate to root pool")
		return
	}

	// Verify the vCon
	verifiedVCon, err := signedVCon.Verify(rootPool)
	if err != nil {
		fmt.Printf("❌ Signature verification failed: %v\n", err)
		return
	}

	fmt.Println("✅ Signature verified successfully!")
	fmt.Printf("Subject: %s\n", verifiedVCon.Subject)
	fmt.Printf("UUID: %s\n", verifiedVCon.UUID)
	fmt.Printf("Created: %s\n", verifiedVCon.CreatedAt)
	fmt.Printf("Number of parties: %d\n", len(verifiedVCon.Parties))
}

// decryptFile decrypts an encrypted vCon file
func decryptFile(path, keyPath, outputPath string) {
	fmt.Printf("Decrypting %s...\n", path)

	// Read the vCon file
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		return
	}

	var encryptedVCon vcon.EncryptedVCon
	if err := json.Unmarshal(data, &encryptedVCon); err != nil {
		fmt.Printf("❌ Error parsing encrypted vCon: %v\n", err)
		return
	}

	// Read the private key
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		fmt.Printf("❌ Error reading private key: %v\n", err)
		return
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY" {
		fmt.Println("❌ Failed to decode PEM block containing private key")
		return
	}

	var privateKey *rsa.PrivateKey
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		var parsedKey interface{}
		parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			fmt.Printf("❌ Error parsing private key: %v\n", err)
			return
		}

		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			fmt.Println("❌ Private key is not an RSA key")
			return
		}
	}

	// Decrypt the vCon
	decrypted, err := encryptedVCon.Decrypt(privateKey)
	if err != nil {
		fmt.Printf("❌ Error decrypting vCon: %v\n", err)
		return
	}

	// Marshal the decrypted vCon
	decryptedData, err := json.MarshalIndent(decrypted, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling decrypted vCon: %v\n", err)
		return
	}

	// Determine output path
	if outputPath == "" {
		ext := filepath.Ext(path)
		outputPath = path[:len(path)-len(ext)] + ".decrypted" + ext
	}

	// Write the decrypted vCon
	if err := os.WriteFile(outputPath, decryptedData, 0644); err != nil {
		fmt.Printf("❌ Error writing decrypted vCon: %v\n", err)
		return
	}

	fmt.Printf("✅ Decrypted vCon written to %s\n", outputPath)
}
