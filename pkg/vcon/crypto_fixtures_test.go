package vcon_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fixtures loads the sample leaf key / leaf certificate / root CA that
// live under testdata/keys/ and returns:
//
//   * rsa.PrivateKey  – the signer (leaf.key)
//   * *x509.Certificate – the leaf certificate (leaf.crt)
//   * *x509.CertPool    – a pool that contains the root CA (root.crt)
//
// It calls t.Fatalf on any error so the caller doesn’t have to.
func loadKeys(t *testing.T) (*rsa.PrivateKey, *x509.Certificate, *x509.CertPool) {
	t.Helper()

	// ---- locate testdata/keys relative to this file ----
	_, thisFile, _, _ := runtime.Caller(0) // nolint: dogsled
	keyDir := filepath.Join(filepath.Dir(thisFile), "testdata", "keys")

	readPEM := func(name string) *pem.Block {
		raw, err := os.ReadFile(filepath.Join(keyDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		b, _ := pem.Decode(raw)
		if b == nil {
			t.Fatalf("decode %s: no PEM block", name)
		}
		return b
	}

	// ---- parse leaf private key ----
	leafKeyBlock := readPEM("leaf.key")
	var leafKey *rsa.PrivateKey
	switch leafKeyBlock.Type {
	case "RSA PRIVATE KEY":
		k, err := x509.ParsePKCS1PrivateKey(leafKeyBlock.Bytes)
		if err != nil {
			t.Fatalf("parse leaf.key: %v", err)
		}
		leafKey = k
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(leafKeyBlock.Bytes)
		if err != nil {
			t.Fatalf("parse leaf.key: %v", err)
		}
		var ok bool
		leafKey, ok = k.(*rsa.PrivateKey)
		if !ok {
			t.Fatalf("leaf.key is not an RSA key")
		}
	default:
		t.Fatalf("leaf.key: unsupported PEM type %q", leafKeyBlock.Type)
	}

	// ---- parse leaf certificate ----
	leafCertBlock := readPEM("leaf.crt")
	leafCert, err := x509.ParseCertificate(leafCertBlock.Bytes)
	if err != nil {
		t.Fatalf("parse leaf.crt: %v", err)
	}

	// ---- load root CA into a pool ----
	rootPool := x509.NewCertPool()
	rootPEM, err := os.ReadFile(filepath.Join(keyDir, "root.crt"))
	if err != nil {
		t.Fatalf("read root.crt: %v", err)
	}
	if !rootPool.AppendCertsFromPEM(rootPEM) {
		t.Fatalf("append root.crt to pool: failed")
	}

	return leafKey, leafCert, rootPool
}
