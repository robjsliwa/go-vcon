package vcon

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

// SignedVCon wraps a signed container.
type SignedVCon struct {
	JWS string `json:"jws"`
}

// EncryptedVCon wraps an encrypted container.
type EncryptedVCon struct {
	JWE string `json:"jwe"`
}

// Sign serializes v and returns a detached-payload JWS (General JSON).
func (v *VCon) Sign(signer crypto.Signer, certs []*x509.Certificate) (*SignedVCon, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Create a SigningKey with the provided signer
	signingKey := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       signer,
	}

	// Set up the signer options
	opts := &jose.SignerOptions{}

	// Add the X.509 certificate chain to the JWS header if provided
	if len(certs) > 0 {
		// Convert raw certificate data for X5c header
		var rawCerts []string
		for _, cert := range certs {
			// X5c expects base64 standard encoded certs (not raw bytes)
			rawCerts = append(rawCerts, base64.StdEncoding.EncodeToString(cert.Raw))
		}

		// Set the X5c header with properly encoded certificate data
		opts = opts.WithHeader(jose.HeaderKey("x5c"), rawCerts)
	}

	// Create the signer
	j, err := jose.NewSigner(signingKey, opts)
	if err != nil {
		return nil, err
	}

	// Sign the payload
	jws, err := j.Sign(payload)
	if err != nil {
		return nil, err
	}

	// Serialize the JWS
	serialised, err := jws.CompactSerialize()
	if err != nil {
		return nil, err
	}
	return &SignedVCon{JWS: serialised}, nil
}

// Verify checks signature and returns the inner VCon.
func (sv *SignedVCon) Verify(rootPool *x509.CertPool) (*VCon, error) {
	jws, err := jose.ParseSigned(sv.JWS, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return nil, err
	}

	// Extract certificate chain from x5c header
	if len(jws.Signatures) == 0 {
		return nil, errors.New("no signature found in JWS")
	}

	sig := jws.Signatures[0]
	x5cHeader, ok := sig.Protected.ExtraHeaders[jose.HeaderKey("x5c")]
	if !ok {
		return nil, errors.New("no x5c header in JWS")
	}

	// Parse the certificate chain
	var certChain []*x509.Certificate
	x5cData, ok := x5cHeader.([]interface{})
	if !ok {
		return nil, errors.New("invalid x5c header format")
	}

	for _, certData := range x5cData {
		// Handle different possible formats of certificate data
		var certBytes []byte

		switch cd := certData.(type) {
		case []byte:
			certBytes = cd
		case string:
			var err error
			certBytes, err = base64.StdEncoding.DecodeString(cd)
			if err != nil {
				return nil, fmt.Errorf("failed to decode certificate: %w", err)
			}
		default:
			return nil, errors.New("invalid certificate format in x5c")
		}

		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		certChain = append(certChain, cert)
	}

	if len(certChain) == 0 {
		return nil, errors.New("no certificates found in x5c header")
	}

	// Verify the certificate chain
	leafCert := certChain[0]
	intermediates := x509.NewCertPool()
	for _, cert := range certChain[1:] {
		intermediates.AddCert(cert)
	}

	verifyOpts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediates,
	}

	if _, err := leafCert.Verify(verifyOpts); err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	// Verify the signature using the leaf certificate's public key
	payload, err := jws.Verify(leafCert.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	var v VCon
	if err = json.Unmarshal(payload, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Encrypt encrypts a SignedVCon for the given recipients.
func (sv *SignedVCon) Encrypt(recipients []jose.Recipient) (*EncryptedVCon, error) {
	enc, err := jose.NewMultiEncrypter(jose.A256GCM, recipients, nil)
	if err != nil {
		return nil, err
	}
	jwe, err := enc.Encrypt([]byte(sv.JWS))
	if err != nil {
		return nil, err
	}
	s, _ := jwe.CompactSerialize()
	return &EncryptedVCon{JWE: s}, nil
}

func (ev *EncryptedVCon) Decrypt(key any) (*SignedVCon, error) {
	// Define allowed key algorithms
	keyAlgs := []jose.KeyAlgorithm{
		jose.RSA_OAEP,
		jose.RSA_OAEP_256,
		jose.ECDH_ES,
		jose.ECDH_ES_A128KW,
		jose.ECDH_ES_A192KW,
		jose.ECDH_ES_A256KW,
	}

	// Define allowed content encryption algorithms
	encAlgs := []jose.ContentEncryption{
		jose.A256GCM,
	}

	jwe, err := jose.ParseEncrypted(ev.JWE, keyAlgs, encAlgs)
	if err != nil {
		return nil, err
	}

	decrypted, err := jwe.Decrypt(key)
	if err != nil {
		return nil, err
	}

	// Validate that the decrypted content is a valid JWS
	signedVCon := &SignedVCon{JWS: string(decrypted)}
	if _, err := jose.ParseSigned(signedVCon.JWS, []jose.SignatureAlgorithm{jose.RS256}); err != nil {
		return nil, fmt.Errorf("decrypted content is not a valid JWS: %w", err)
	}

	return signedVCon, nil
}
