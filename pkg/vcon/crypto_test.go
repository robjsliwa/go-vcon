package vcon_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
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

// TestSignAndVerify tests signing and verification of a vCon
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

	// Sign the vCon
	signed, err := v.Sign(privateKey, certs)
	require.NoError(t, err)
	assert.NotNil(t, signed)

	// Verify the signed vCon using the built-in Verify method
	verified, err := signed.Verify(rootPool)
	require.NoError(t, err)

	// Check that the verification succeeded and returned the original content
	assert.Equal(t, v.Subject, verified.Subject)
	assert.Equal(t, v.UUID, verified.UUID)
	assert.Equal(t, len(v.Parties), len(verified.Parties))
	assert.Equal(t, v.Parties[0].Name, verified.Parties[0].Name)
}

// TestEncryptAndDecrypt tests encryption and decryption of a signed vCon
func TestEncryptAndDecrypt(t *testing.T) {
	// Generate a test certificate
	privateKey, certs, err := generateTestCertificate()
	require.NoError(t, err)

	// Create a test vCon
	v := vcon.New()
	v.Subject = "Test vCon"
	v.AddParty(vcon.Party{Name: "Test Person"})

	// Sign the vCon
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

	// Debug: show what decrypted looks like
	decryptedJSON, err := json.MarshalIndent(decrypted, "", "  ")
	require.NoError(t, err)
	t.Logf("Decrypted structure: %s", string(decryptedJSON))

	// The decrypted result should be the original signed JWS structure
	// Convert back to SignedVCon to verify
	var signedVCon vcon.SignedVCon
	signedVCon.JSON = decrypted
	
	// Create a root pool for verification
	rootPool := x509.NewCertPool()
	rootPool.AddCert(certs[0])

	// Verify the decrypted content
	verifiedAfterDecrypt, err := signedVCon.Verify(rootPool)
	require.NoError(t, err)

	// Check that we got back our original content
	assert.Equal(t, v.Subject, verifiedAfterDecrypt.Subject)
	assert.Equal(t, v.UUID, verifiedAfterDecrypt.UUID)
	assert.Equal(t, len(v.Parties), len(verifiedAfterDecrypt.Parties))
	assert.Equal(t, v.Parties[0].Name, verifiedAfterDecrypt.Parties[0].Name)
}

// TestCompleteRoundTrip tests the complete vcon->sign->encrypt->decrypt->verify->original vcon flow
func TestCompleteRoundTrip(t *testing.T) {
	// Generate a test certificate
	privateKey, certs, err := generateTestCertificate()
	require.NoError(t, err)

	// Create a root pool for verification
	rootPool := x509.NewCertPool()
	rootPool.AddCert(certs[0])

	// Step 1: Create original vCon
	original := vcon.New()
	original.Subject = "Complete Round Trip Test"
	partyIdx := original.AddParty(vcon.Party{
		Name: "Alice Smith",
		Tel:  "tel:+12025551234",
		Role: "customer",
	})
	
	// Add some dialog to make it more realistic
	now := time.Now().UTC()
	original.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &now,
		Duration:   120.5,
		Parties:    []int{partyIdx},
		Originator: partyIdx,
		MediaType:  "audio/wav",
		Body:       "base64encodedaudiodata...",
		Encoding:   "base64",
	})

	// Validate original
	err = original.Validate()
	require.NoError(t, err, "Original vCon should be valid")

	// Step 2: Sign the vCon
	signed, err := original.Sign(privateKey, certs)
	require.NoError(t, err, "Signing should succeed")
	assert.NotNil(t, signed)

	// Step 3: Encrypt the signed vCon
	recipient := jose.Recipient{
		Algorithm: jose.RSA_OAEP,
		Key:       &privateKey.PublicKey,
	}
	encrypted, err := signed.Encrypt([]jose.Recipient{recipient})
	require.NoError(t, err, "Encryption should succeed")
	assert.NotNil(t, encrypted)

	// Step 4: Decrypt the encrypted vCon
	decrypted, err := encrypted.Decrypt(privateKey)
	require.NoError(t, err, "Decryption should succeed")
	assert.NotNil(t, decrypted)

	// Step 5: Convert decrypted result back to SignedVCon
	// The decrypted result is the original JWS structure
	signedAfterDecrypt := vcon.SignedVCon{JSON: decrypted}

	// Step 6: Verify the signature
	finalVCon, err := signedAfterDecrypt.Verify(rootPool)
	require.NoError(t, err, "Signature verification should succeed")

	// Step 7: Verify all content matches the original
	assert.Equal(t, original.Subject, finalVCon.Subject, "Subject should match")
	assert.Equal(t, original.UUID, finalVCon.UUID, "UUID should match")
	assert.Equal(t, original.Vcon, finalVCon.Vcon, "Version should match")
	assert.Equal(t, len(original.Parties), len(finalVCon.Parties), "Party count should match")
	assert.Equal(t, len(original.Dialog), len(finalVCon.Dialog), "Dialog count should match")
	
	// Check party details
	assert.Equal(t, original.Parties[0].Name, finalVCon.Parties[0].Name, "Party name should match")
	assert.Equal(t, original.Parties[0].Tel, finalVCon.Parties[0].Tel, "Party tel should match")
	assert.Equal(t, original.Parties[0].Role, finalVCon.Parties[0].Role, "Party role should match")
	
	// Check dialog details
	assert.Equal(t, original.Dialog[0].Type, finalVCon.Dialog[0].Type, "Dialog type should match")
	assert.Equal(t, original.Dialog[0].Duration, finalVCon.Dialog[0].Duration, "Dialog duration should match")
	assert.Equal(t, original.Dialog[0].MediaType, finalVCon.Dialog[0].MediaType, "Dialog media type should match")
	assert.Equal(t, original.Dialog[0].Body, finalVCon.Dialog[0].Body, "Dialog body should match")

	// Final validation
	err = finalVCon.Validate()
	require.NoError(t, err, "Final vCon should still be valid")

	t.Log("✅ Complete round trip test successful: vcon->sign->encrypt->decrypt->verify->original vcon")
}

// TestVerifyRoundTrip uses the test key fixtures for verification
func TestVerifyRoundTrip(t *testing.T) {
    leafKey, leafCert, rootPool := loadKeys(t) // helper parses PEM files

    vc := vcon.New()
    vc.Subject = "Test with fixture keys"

    signed, err := vc.Sign(leafKey, []*x509.Certificate{leafCert})
    if err != nil { t.Fatalf("sign: %v", err) }

    got, err := signed.Verify(rootPool)
    if err != nil { t.Fatalf("verify: %v", err) }

    if got.Subject != vc.Subject {
        t.Fatalf("subject mismatch: want %s got %s", vc.Subject, got.Subject)
    }
    
    // Additional checks
    assert.Equal(t, vc.UUID, got.UUID, "UUID should match")
    assert.Equal(t, vc.Vcon, got.Vcon, "Version should match")
}
