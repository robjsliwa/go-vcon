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
		E164: "+12025551234",
	})
	
	v.AddParty(vcon.Party{
		Name: "Jane Smith",
		E164: "+12025555678",
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

