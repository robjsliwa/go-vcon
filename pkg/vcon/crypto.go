package vcon

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

// SignedVCon wraps a signed container.
type SignedVCon struct {
	JSON map[string]any `json:"jws"`
}

// EncryptedVCon wraps an encrypted container.
type EncryptedVCon struct {
	JSON map[string]any `json:"jwe"`
}

// Sign generates a General‑JSON JWS with detached payload.
func (v *VCon) Sign(signer crypto.Signer, chain []*x509.Certificate) (*SignedVCon, error) {
	payload, err := Canonicalise(v)
	if err != nil { return nil, err }

	// embed x5c
	var x5c []string
	for _, c := range chain {
		x5c = append(x5c, base64.StdEncoding.EncodeToString(c.Raw))
	}

	j, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: signer},
		(&jose.SignerOptions{}).WithHeader("x5c", x5c).WithHeader("uuid", v.UUID))
	if err != nil { return nil, err }
	obj, err := j.Sign(payload)
	if err != nil { return nil, err }

	general := obj.FullSerialize()
	var gen map[string]any
	if err = json.Unmarshal([]byte(general), &gen); err != nil { return nil, err }
	gen["payload"] = base64.RawURLEncoding.EncodeToString(payload)

	return &SignedVCon{JSON: gen}, nil
}

// Verify validates all signatures, certificate chains and canonicalization.
// On success it returns the decoded VCon.
func (sv *SignedVCon) Verify(rootPool *x509.CertPool) (*VCon, error) {
	raw, err := json.Marshal(sv.JSON)
	if err != nil {
		return nil, fmt.Errorf("marshal signed object: %w", err)
	}

	jws, err := jose.ParseSigned(string(raw), []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return nil, fmt.Errorf("parse JWS: %w", err)
	}

	var (
		refPayload []byte // canonical payload after first successful sig
		vc         *VCon  // decoded vCon to return
	)

	for idx, sig := range jws.Signatures {
		// 2.a validate and extract x5c chain
		chains, err := sig.Header.Certificates(x509.VerifyOptions{Roots: rootPool})
		if err != nil {
			return nil, fmt.Errorf("sig[%d] bad cert chain: %w", idx, err)
		}
		leaf := chains[0][0] // leaf cert is first in verified chain

		// 2.b verify signature with leaf’s public key
		payload, err := jws.Verify(leaf.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("sig[%d] signature invalid: %w", idx, err)
		}

		if idx == 0 {
			refPayload = payload

			var v VCon
			if err := json.Unmarshal(payload, &v); err != nil {
				return nil, fmt.Errorf("decode vCon: %w", err)
			}

			canon, _ := Canonicalise(&v)
			if !bytes.Equal(canon, payload) {
				return nil, errors.New("payload not RFC 8785 canonical")
			}

			if hu, ok := sig.Header.ExtraHeaders["uuid"].(string); ok && hu != v.UUID {
				return nil, errors.New("header uuid ≠ body uuid")
			}

			vc = &v
		} else {
			if !bytes.Equal(refPayload, payload) {
				return nil, fmt.Errorf("sig[%d] payload mismatch", idx)
			}
		}
	}

	if vc == nil {
		return nil, errors.New("no valid signatures")
	}
	return vc, nil
}


// Encrypt turns a *signed* vCon (General-JSON JWS in sv.JSON) into a
// complete-serialization JWE.
func (sv *SignedVCon) Encrypt(rcpts []jose.Recipient) (*EncryptedVCon, error) {
	if len(rcpts) == 0 {
		return nil, errors.New("no recipients supplied")
	}

	plain, err := Canonicalise(sv.JSON)
	if err != nil {
		return nil, fmt.Errorf("canonicalise signed vCon: %w", err)
	}

	var tmp struct{ UUID string `json:"uuid"` }
	if err := json.Unmarshal(plain, &tmp); err != nil {
		return nil, fmt.Errorf("extract uuid: %w", err)
	}

	opts := (&jose.EncrypterOptions{}).
		// typ & cty aren’t strictly required but useful for tooling
		WithType("vcon+jwe").
		WithContentType("application/vcon+json").
		WithHeader("uuid", tmp.UUID)

	enc, err := jose.NewMultiEncrypter(jose.A256CBC_HS512, rcpts, opts)
	if err != nil {
		return nil, fmt.Errorf("new encrypter: %w", err)
	}

	jweObj, err := enc.Encrypt(plain)
	if err != nil {
		return nil, fmt.Errorf("encrypt vCon: %w", err)
	}

	var jweMap map[string]any
	if err := json.Unmarshal([]byte(jweObj.FullSerialize()), &jweMap); err != nil {
		return nil, fmt.Errorf("unmarshal JWE: %w", err)
	}
	jweMap["unprotected"] = map[string]any{
		"uuid": tmp.UUID,
		"cty":  "application/vcon+json",
		"enc":  string(jose.A256CBC_HS512),
	}

	return &EncryptedVCon{JSON: jweMap}, nil
}

// Decrypt unwraps the JWE using the supplied **private RSA key**.
// It returns the plaintext object as a generic map.
func (ev *EncryptedVCon) Decrypt(priv *rsa.PrivateKey) (map[string]any, error) {
	raw, err := json.Marshal(ev.JSON)
	if err != nil {
		return nil, fmt.Errorf("marshal JWE: %w", err)
	}

	jweObj, err := jose.ParseEncrypted(
		string(raw),
		[]jose.KeyAlgorithm{jose.RSA_OAEP, jose.RSA_OAEP_256},
		[]jose.ContentEncryption{jose.A256CBC_HS512},
	)
	if err != nil {
		return nil, fmt.Errorf("parse JWE: %w", err)
	}

	plain, err := jweObj.Decrypt(priv)
	if err != nil {
		return nil, fmt.Errorf("decrypt JWE: %w", err)
	}

	var out map[string]any
	if err := json.Unmarshal(plain, &out); err != nil {
		return nil, fmt.Errorf("decode plaintext: %w", err)
	}
	return out, nil
}
