package vcon_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestCertificate creates a self-signed certificate for testing
func generateTestCertificate() (*rsa.PrivateKey, []*x509.Certificate, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Create certificate template
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(24 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "test.example.com",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, []*x509.Certificate{cert}, nil
}

func TestSignAndVerify(t *testing.T) {
	// Generate a test certificate
	privateKey, certs, err := generateTestCertificate()
	require.NoError(t, err)
	
	// Create a root pool for verification
	rootPool := x509.NewCertPool()
	rootPool.AddCert(certs[0])
	
	// Create a test vCon
	v := vcon.New()
	v.Subject = "Test vCon"
	v.AddParty(vcon.Party{Name: "Test Person"})
	
	// Sign the vCon - directly pass certificates to the Sign method
	signed, err := v.Sign(privateKey, certs)
	require.NoError(t, err)
	assert.NotNil(t, signed)
	
	// Verify the signed vCon
	verified, err := signed.Verify(rootPool)
	require.NoError(t, err)
	
	// Check that the verification succeeded and returned the original content
	assert.Equal(t, v.Subject, verified.Subject)
	assert.Equal(t, v.UUID, verified.UUID)
	assert.Equal(t, len(v.Parties), len(verified.Parties))
}

func TestEncryptAndDecrypt(t *testing.T) {
	// Generate a test certificate
	privateKey, certs, err := generateTestCertificate()
	require.NoError(t, err)
	
	// Create a test vCon
	v := vcon.New()
	v.Subject = "Test vCon"
	v.AddParty(vcon.Party{Name: "Test Person"})
	
	// Sign the vCon first - directly pass certificates
	signed, err := v.Sign(privateKey, certs)
	require.NoError(t, err)
	
	// Create recipient for encryption
	recipient := jose.Recipient{
		Algorithm: jose.RSA_OAEP,
		Key:       &privateKey.PublicKey,
	}

	// Encrypt the signed vCon
	encrypted, err := signed.Encrypt([]jose.Recipient{recipient})
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	
	// Decrypt the encrypted vCon
	decrypted, err := encrypted.Decrypt(privateKey)
	require.NoError(t, err)
	assert.NotNil(t, decrypted)
	
	// Create a root pool for verification
	rootPool := x509.NewCertPool()
	rootPool.AddCert(certs[0])
	
	// Verify the decryption succeeded
	verified, err := decrypted.Verify(rootPool)
	require.NoError(t, err)
	
	// Check that we got back our original content
	assert.Equal(t, v.Subject, verified.Subject)
	assert.Equal(t, v.UUID, verified.UUID)
	assert.Equal(t, len(v.Parties), len(verified.Parties))
}
