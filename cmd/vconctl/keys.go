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
	"time"

	"github.com/spf13/cobra"
)

// Command: genkey

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

// helper utils

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
