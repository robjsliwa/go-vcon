# go-vcon

A Go implementation of the vCon (Virtual Conversation) container specification based on [draft-ietf-vcon-vcon-container-03](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-container/).

## Overview

This library provides functionality for working with vCon containers in Go, including:

- Creation and modification of vCon containers
- Validation against the vCon JSON schema
- Cryptographic operations (signing, verification, encryption, decryption)
- File reference handling

The project also includes a command-line tool (`vconctl`) for performing common vCon operations.

## Installation

### Library

```bash
go get github.com/robjsliwa/go-vcon
```

### Command-line Tool

```bash
go install github.com/robjsliwa/go-vcon/cmd/vconctl@latest
```

Or build from source:

```bash
git clone https://github.com/robjsliwa/go-vcon.git
cd go-vcon
go build -o vconctl ./cmd/vconctl
```

## Usage

### Library Usage

```go
package main

import (
	"fmt"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
)

func main() {
	// Create a new vCon
	v := vcon.New()
	
	// Set basic properties
	v.Subject = "Sample conversation"
	
	// Add parties
	v.AddParty(vcon.Party{
		Name: "John Doe",
		Tel: "tel:+12025551234",
	})
	
	v.AddParty(vcon.Party{
		Name: "Jane Smith",
		Tel: "tel:+12025555678",
	})
	
	// Validate
	if err := v.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}
	
	fmt.Printf("Created valid vCon with UUID: %s\n", v.UUID)
}
```

### Command-line Usage

The `vconctl` tool provides several commands for working with vCon files:

```
vconctl is a command line utility for validating, signing, encrypting,
verifying, and decrypting vCon (Virtual Conversation) files.

Usage:
  vconctl [command]

Available Commands:
  decrypt     Decrypt an encrypted vCon file
  encrypt     Encrypt a vCon file
  help        Help about any command
  sign        Sign a vCon file
  validate    Validate a vCon file
  verify      Verify a signed vCon file
```

#### Validate a vCon

```bash
vconctl validate --file sample.json
```

#### Sign a vCon

```bash
vconctl sign --file sample.json --cert certificate.pem --key private-key.pem --out signed.json
```

#### Verify a Signed vCon

```bash
vconctl verify --file signed.json --cert certificate.pem
```

#### Encrypt a Signed vCon

```bash
vconctl encrypt --file signed.json --cert certificate.pem --out encrypted.json
```

#### Decrypt an Encrypted vCon

```bash
vconctl decrypt --file encrypted.json --key private-key.pem --out decrypted.json
```

### Sample vCon Files

Below are sample vCon files you can use for testing and demonstration purposes.

#### Simple vCon

Save this to a file named `simple-vcon.json`:

```json
{
  "vcon": "0.0.3",
  "uuid": "9b583dd6-31b2-4403-b74e-271f45f97ada",
  "created_at": "2023-06-15T14:25:33Z",
  "subject": "Customer Support Call",
  "parties": [
    {
      "name": "John Doe",
      "e164": "+12025551234",
      "role": "customer"
    },
    {
      "name": "Jane Smith",
      "e164": "+18005559876",
      "role": "agent"
    }
  ]
}
```

#### Comprehensive vCon

Save this to a file named `comprehensive-vcon.json`:

```json
{
  "vcon": "0.0.3",
  "uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "created_at": "2023-06-15T14:30:00Z",
  "updated_at": "2023-06-15T15:45:12Z",
  "subject": "Technical Support - Network Connectivity Issue",
  "parties": [
    {
      "name": "Bob Johnson",
      "e164": "+12025551111",
      "role": "customer",
      "addr": "bob.johnson@example.com",
      "loc": "New York, USA"
    },
    {
      "name": "Sarah Lee",
      "e164": "+18005552222",
      "role": "support",
      "addr": "support@company.example.com"
    },
    {
      "name": "AI Assistant",
      "role": "virtual-agent",
      "addr": "ai-assistant@company.example.com"
    }
  ],
  "dialog": [
    {
      "start": "2023-06-15T14:30:00Z",
      "duration": 500,
      "media_type": "audio/mp3",
      "orig_party": 0,
      "dest_parties": [1],
      "content": {
        "url": "https://example.com/recordings/call-1234.mp3",
        "content_hash": "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
      },
      "interaction_id": "call-1234",
      "type": "text",
      "body": "Bob Johnson calls for technical support regarding a network connectivity issue.",
      "encoding": "base64",
      "parties": 1
    },
    {
      "start": "2023-06-15T14:35:30Z",
      "duration": 500,
      "media_type": "text/plain",
      "orig_party": 1,
      "dest_parties": [0],
      "content": {
        "body": "I've analyzed your network logs and found the issue. Your router firmware needs to be updated to the latest version.",
        "encoding": "base64"
      },
      "interaction_id": "chat-5678",
      "type": "text",
      "body": "Jane Smith responds with a solution for the network issue.",
      "encoding": "base64",
      "parties": 1
    }
  ],
  "analysis": [
    {
      "type": "sentiment",
      "vendor": "AIAnalytics, Inc.",
      "product": "SentimentAnalyzer v2.1",
      "content": {
        "body": "eyJjdXN0b21lcl9zZW50aW1lbnQiOiAicG9zaXRpdmUiLCAiYWdlbnRfcGVyZm9ybWFuY2UiOiAiZXhjZWxsZW50In0=",
        "encoding": "base64"
      }
    },
    {
      "type": "transcription",
      "vendor": "Transcribe Pro",
      "product": "AutoTranscribe v3.0",
      "content": {
        "url": "https://example.com/transcripts/call-1234.txt",
        "content_hash": "sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3"
      }
    }
  ],
  "attachments": [
    {
      "body": "eyJuZXR3b3JrX2xvZ3MiOiAidHJ1bmNhdGVkIGZvciByZWFkYWJpbGl0eSJ9",
      "encoding": "base64",
      "party": 0,
      "start": "2023-06-15T14:30:00Z"
    }
  ]
}
```

You can use these sample files to test the vconctl commands:

```bash
# Validate
vconctl validate simple-vcon.json
vconctl validate comprehensive-vcon.json

# Generate test key and certificate for signing/encryption
vconctl genkey

# Sign
vconctl sign simple-vcon.json --cert test_cert.pem --key test_key.pem

# Verify
vconctl verify simple-vcon.signed.json --cert test_cert.pem

# Encrypt
vconctl encrypt simple-vcon.signed.json --cert test_cert.pem

# Decrypt
vconctl decrypt simple-vcon.signed.encrypted.json --key test_key.pem
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [IETF vCon Working Group](https://datatracker.ietf.org/wg/vcon/about/)
- [draft-ietf-vcon-vcon-container-03](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-container/)

