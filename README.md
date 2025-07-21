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
go get -u github.com/robjsliwa/go-vcon
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

## Development

### Running Tests

To run the test suite:

```bash
go test ./...
```

### Test Coverage

To run tests with code coverage:

```bash
# Run tests with coverage
go test -cover ./...

# Generate a coverage profile
go test -coverprofile=coverage.out ./...

# Display coverage in the browser
go tool cover -html=coverage.out

# Get a coverage summary in the terminal
go tool cover -func=coverage.out
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
	// Create a new vCon with domain for UUID generation
	v := vcon.New("example.com")
	
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
vconctl is a command-line utility for validating, signing, encrypting, verifying, and decrypting vCon (Virtual Conversation) files.

Usage:
  vconctl [command]

Available Commands:
  convert     Convert external artifacts (audio, zoom, email) into vCon containers
  decrypt     Decrypt an encrypted vCon file
  encrypt     Encrypt a signed vCon for one recipient
  genkey      Generate a test RSA key pair and self-signed certificate
  help        Help about any command
  sign        Sign a vCon file using a private key and certificate
  validate    Validate a vCon file
  verify      Verify the signature on a signed vCon
```

### Command Options

#### validate
```bash
vconctl validate [file1] [file2] ...
```
- Validates one or more vCon files against the JSON schema
- Returns ✅ for valid files, ❌ with error details for invalid files

#### genkey
```bash
vconctl genkey [flags]
```
Flags:
- `--key, -k`: Output private-key path (default: test_key.pem)
- `--cert, -c`: Output certificate path (default: test_cert.pem)

#### sign
```bash
vconctl sign [file] [flags]
```
Flags:
- `--key, -k`: Path to private key file (required)
- `--cert, -c`: Path to certificate file (required)
- `--output, -o`: Path to output file (defaults to `<file>.signed.json`)

#### verify
```bash
vconctl verify [file] [flags]
```
Flags:
- `--cert, -c`: Path to trust anchor (leaf or CA) (required)

#### encrypt
```bash
vconctl encrypt [file] [flags]
```
Flags:
- `--cert, -c`: Path to recipient certificate (required)
- `--output, -o`: Path to output file (defaults to `<file>.encrypted.json`)

#### decrypt
```bash
vconctl decrypt [file] [flags]
```
Flags:
- `--key, -k`: Path to private key file (required)
- `--output, -o`: Path to output file (defaults to `<file>.decrypted.json`)

#### convert

The convert command allows you to create vCon files from various external sources:

##### convert audio
```bash
vconctl convert audio --input <file|url> --party <spec> [--party <spec> ...] [flags]
```
Flags:
- `--input`: Path or URL to recording file (required)
- `--party`: Party specification in format 'name,tel:+1555...' or 'name,mailto:user@domain'
- `--date`: Recording start time in RFC3339 format (default: file modification time)
- `--output, -o`: Output vCon file path (default: `<input_name>.vcon.json`)
- `--domain`: Domain name for UUID generation (default: vcon.example.com)

Example:
```bash
vconctl convert audio --input recording.wav --party "Alice,tel:+12025551234" --party "Bob,tel:+12025555678" --date "2025-07-20T23:20:50.52Z" --domain example.com -o conversation.vcon.json
```

```bash
go run ./cmd/vconctl convert audio --input https://raw.githubusercontent.com/robjsliwa/go-vcon/main/testdata/sample_vcons/1745501752.21.wav --party rob,tel:+12151235555 --party alice,tel:+12671235555 --date 2025-07-20T23:20:50.52Z --domain example.com -o testdata/sample_vcons/rec2.vcon.json
```

##### convert zoom
```bash
vconctl convert zoom <folder>
```
Converts a Zoom recording folder into a vCon file. The folder should contain Zoom recording files and metadata.

Example:
```bash
vconctl convert zoom ./zoom_meeting_folder
```

##### convert email
```bash
vconctl convert email <file.eml>
```
Converts an RFC-822 email message file into a vCon file.

Example:
```bash
vconctl convert email message.eml
```

#### Global Flags

All commands support these global flags:
- `--domain string`: Domain name for UUID generation (default "vcon.example.com")

### Command Examples

#### Validate a vCon

```bash
vconctl validate simple-vcon.json
```

#### Generate Test Keys and Certificates

Before signing or encrypting vCon files, you'll need a key pair and certificate. You can generate test ones:

```bash
vconctl genkey
```

This will create `test_key.pem` (private key) and `test_cert.pem` (certificate) in the current directory. You can specify custom paths:

```bash
vconctl genkey --key my_private_key.pem --cert my_certificate.pem
```

#### Sign a vCon

```bash
vconctl sign simple-vcon.json --key test_key.pem --cert test_cert.pem
```

This creates `simple-vcon.signed.json`. You can specify a custom output file:

```bash
vconctl sign simple-vcon.json --key test_key.pem --cert test_cert.pem --output my_signed_vcon.json
```

#### Verify a Signed vCon

```bash
vconctl verify simple-vcon.signed.json --cert test_cert.pem
```

#### Encrypt a Signed vCon

```bash
vconctl encrypt simple-vcon.signed.json --cert test_cert.pem
```

This creates `simple-vcon.signed.encrypted.json`. You can specify a custom output file:

```bash
vconctl encrypt simple-vcon.signed.json --cert test_cert.pem --output my_encrypted_vcon.json
```

#### Decrypt an Encrypted vCon

```bash
vconctl decrypt simple-vcon.signed.encrypted.json --key test_key.pem
```

This creates `simple-vcon.signed.encrypted.decrypted.json`. You can specify a custom output file:

```bash
vconctl decrypt simple-vcon.signed.encrypted.json --key test_key.pem --output my_decrypted_vcon.json
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
      "tel": "+12025551234",
      "role": "customer"
    },
    {
      "name": "Jane Smith",
      "tel": "+18005559876",
      "role": "agent"
    }
  ]
}
```

When you run validation on this vCon file, it will pass successfully:

```bash
vconctl validate simple-vcon.json
Validating simple-vcon.json…
✅ simple-vcon.json is valid
```

#### Comprehensive vCon - Fails Validation

Save this to a file named `comprehensive-vcon-errors.json`:

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
When you run validation on this vCon file, it will fail due to missing required fields and incorrect data types:

```bash
vconctl validate comprehensive-vcon-errors.json
Validating comprehensive-vcon-errors.json…
❌ Error: schema validation failed: jsonschema validation failed with 'file:///Users/robertsliwa/Documents/projects/tmp/robjsliwa/go-vcon/vcon.schema.json#'
- at '/parties': validation failed
  - at '/parties/0': additional properties 'loc', 'e164', 'addr' not allowed
  - at '/parties/1': additional properties 'e164', 'addr' not allowed
  - at '/parties/2': additional properties 'addr' not allowed
- at '/dialog': validation failed
  - at '/dialog/0': additional properties 'dest_parties', 'content', 'media_type', 'orig_party' not allowed
  - at '/dialog/1': additional properties 'media_type', 'orig_party', 'dest_parties', 'content' not allowed
- at '/analysis': validation failed
  - at '/analysis/0': additional properties 'content' not allowed
  - at '/analysis/1': additional properties 'content' not allowed
- at '/attachments/0/blah': false schema
```

#### Comprehensive vCon

Save this to a file named `comprehensive-vcon.json`:

```bash
{
  "vcon": "0.0.3",
  "uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "created_at": "2023-06-15T14:30:00Z",
  "updated_at": "2023-06-15T15:45:12Z",
  "subject": "Technical Support - Network Connectivity Issue",
  "parties": [
    {
      "name": "Bob Johnson",
      "tel": "+12025551111",
      "role": "customer"
    },
    {
      "name": "Sarah Lee",
      "tel": "+18005552222"
    },
    {
      "name": "AI Assistant",
      "role": "virtual-agent"
    }
  ],
  "dialog": [
    {
      "start": "2023-06-15T14:30:00Z",
      "duration": 500,
      "mediatype": "audio/mp3",
      "originator": 0,
      "parties": [1],
      "interaction_id": "call-1234",
      "type": "text",
      "body": "Bob Johnson calls for technical support regarding a network connectivity issue.",
      "encoding": "base64"
    },
    {
      "start": "2023-06-15T14:35:30Z",
      "duration": 500,
      "mediatype": "text/plain",
      "originator": 0,
      "interaction_id": "chat-5678",
      "type": "text",
      "body": "Jane Smith responds with a solution for the network issue.",
      "encoding": "base64",
      "parties": [1]
    }
  ],
  "analysis": [
    {
      "type": "sentiment",
      "vendor": "AIAnalytics, Inc.",
      "product": "SentimentAnalyzer v2.1",
      "body": "eyJjdXN0b21lcl9zZW50aW1lbnQiOiAicG9zaXRpdmUiLCAiYWdlbnRfcGVyZm9ybWFuY2UiOiAiZXhjZWxsZW50In0=",
      "encoding": "base64"

    },
    {
      "type": "transcription",
      "vendor": "Transcribe Pro",
      "product": "AutoTranscribe v3.0",
      "url": "https://example.com/transcripts/call-1234.txt",
      "content_hash": "sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3"
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

When you run vconctl on this vCon file, it will pass validation:

```bash
vconctl validate comprehensive-vcon.json
Validating comprehensive-vcon.json…
✅ comprehensive-vcon.json is valid
```

#### vconctl Commands

You can use these sample files to test the vconctl commands:

```bash
# Validate
vconctl validate simple-vcon.json
vconctl validate comprehensive-vcon.json

# Generate test key and certificate for signing/encryption
vconctl genkey

# Sign (requires key and certificate)
vconctl sign simple-vcon.json --cert test_cert.pem --key test_key.pem

# Verify (requires certificate/CA)
vconctl verify simple-vcon.signed.json --cert test_cert.pem

# Encrypt (requires certificate for recipient)
vconctl encrypt simple-vcon.signed.json --cert test_cert.pem

# Decrypt (requires private key)
vconctl decrypt simple-vcon.signed.encrypted.json --key test_key.pem

# Convert audio file to vCon
vconctl convert audio --input recording.wav --party "Speaker 1,tel:+12025551234" --party "Speaker 2,tel:+12025555678"

# Convert Zoom recording to vCon
vconctl convert zoom ./zoom_recording_folder

# Convert email to vCon
vconctl convert email message.eml
```

### Complete Workflow Example

Here's a complete example showing the full workflow from validation through encryption and decryption:

```bash
# 1. Validate the original vCon
vconctl validate simple-vcon.json

# 2. Generate test keys and certificate
vconctl genkey

# 3. Sign the vCon
vconctl sign simple-vcon.json --cert test_cert.pem --key test_key.pem
# This creates simple-vcon.signed.json

# 4. Verify the signed vCon
vconctl verify simple-vcon.signed.json --cert test_cert.pem

# 5. Encrypt the signed vCon
vconctl encrypt simple-vcon.signed.json --cert test_cert.pem
# This creates simple-vcon.signed.encrypted.json

# 6. Decrypt the encrypted vCon
vconctl decrypt simple-vcon.signed.encrypted.json --key test_key.pem
# This creates simple-vcon.signed.encrypted.decrypted.json

# 7. Verify the decrypted content (optional)
vconctl verify simple-vcon.signed.encrypted.decrypted.json --cert test_cert.pem

# 8. Convert external sources to vCon (optional examples)
# Convert audio file
vconctl convert audio --input meeting.wav --party "John,tel:+12025551111" --party "Jane,tel:+12025552222" --date "2025-07-20T14:30:00Z"

# Convert Zoom recording
vconctl convert zoom ./zoom_meeting_20250720

# Convert email
vconctl convert email important_conversation.eml
```

### Differences from Python Reference Implementation

While this Go implementation follows the same core vCon specification as the Python reference implementation, there are some differences in the command-line interface:

**Similarities:**
- All core cryptographic operations (sign, verify, encrypt, decrypt)
- JSON schema validation
- Support for the same vCon specification (version 0.0.3)
- Compatible file formats and outputs

**Key Differences:**
- **Simpler command structure**: Go version uses direct commands (`vconctl sign file.json`) vs Python's more complex argument structure
- **Flag-based options**: Go version uses `--key`, `--cert`, `--output` flags instead of positional arguments
- **Automatic output naming**: Go version automatically generates output filenames (e.g., `file.signed.json`) unless specified
- **Convert commands**: Go version includes `convert` subcommands for audio, Zoom, and email import
- **Built-in key generation**: Go version includes `genkey` command for easy test key/certificate generation

**Implemented Features (from Python reference):**
- Core vCon operations (validate, sign, verify, encrypt, decrypt)
- Audio file conversion (`convert audio`)
- Zoom meeting import (`convert zoom`) 
- Email message import (`convert email`)
- Domain-based UUID generation

**Missing Features (compared to Python):**
- Filter plugins system
- HTTP GET/POST operations for remote vCon retrieval/storage
- Advanced analysis features
- Google Meet import functionality
- Comprehensive recording file processing

The Go implementation is designed to be a clean, focused tool for core vCon operations, while the Python implementation provides a more comprehensive toolkit for various vCon workflows.

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

