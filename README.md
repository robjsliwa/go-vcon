# go-vcon

[![Go Reference](https://pkg.go.dev/badge/github.com/robjsliwa/go-vcon.svg)](https://pkg.go.dev/github.com/robjsliwa/go-vcon)
[![Go Report Card](https://goreportcard.com/badge/github.com/robjsliwa/go-vcon)](https://goreportcard.com/report/github.com/robjsliwa/go-vcon)

A Go implementation of the **vCon (Virtual Conversation)** container specification, fully compliant with [draft-ietf-vcon-vcon-core-02](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-core/) (spec version `0.4.0`).

vCon is an IETF standard for encapsulating conversation data -- recordings, transcripts, analysis, and metadata -- into a single, portable JSON container with built-in support for cryptographic signing and encryption.

## Features

- **Create, validate, and manipulate** vCon containers
- **Cryptographic operations** -- JWS signing (RS256) and JWE encryption (RSA-OAEP)
- **JSON Schema validation** against the vCon core specification
- **Extension framework** with a built-in Contact Center (CC) extension per [draft-ietf-vcon-cc-extension-01](https://datatracker.ietf.org/doc/draft-ietf-vcon-cc-extension/)
- **Redaction and amendment** workflows per the specification
- **Form detection** -- identify whether a vCon is unsigned, signed, or encrypted
- **Backward compatibility** -- automatic migration of v0.0.3 vCons to v0.4.0
- **CLI tool** (`vconctl`) for validation, signing, encryption, conversion, and more

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Library Reference](#library-reference)
  - [Creating a vCon](#creating-a-vcon)
  - [Parties](#parties)
  - [Dialogs](#dialogs)
  - [Analysis](#analysis)
  - [Attachments](#attachments)
  - [Validation](#validation)
  - [Signing and Verification](#signing-and-verification)
  - [Encryption and Decryption](#encryption-and-decryption)
  - [Redaction](#redaction)
  - [Amendment](#amendment)
  - [Extensions](#extensions)
  - [Content Hashing](#content-hashing)
  - [Form Detection](#form-detection)
  - [Serialization](#serialization)
- [CLI Reference](#cli-reference)
  - [validate](#validate)
  - [detect](#detect)
  - [genkey](#genkey)
  - [sign](#sign)
  - [verify](#verify)
  - [encrypt](#encrypt)
  - [decrypt](#decrypt)
  - [convert audio](#convert-audio)
  - [convert zoom](#convert-zoom)
  - [convert email](#convert-email)
- [Complete Workflow Examples](#complete-workflow-examples)
- [Sample vCon Files](#sample-vcon-files)
- [Development](#development)
- [Contributing](#contributing)
- [Acknowledgments](#acknowledgments)

---

## Installation

### Library

```bash
go get -u github.com/robjsliwa/go-vcon
```

### CLI Tool

```bash
go install github.com/robjsliwa/go-vcon/cmd/vconctl@latest
```

Or build from source:

```bash
git clone https://github.com/robjsliwa/go-vcon.git
cd go-vcon
go build -o vconctl ./cmd/vconctl
```

### Requirements

- Go 1.24 or later
- `ffprobe` (optional, required only for audio conversion)

---

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/robjsliwa/go-vcon/pkg/vcon"
)

func main() {
    // Create a new vCon
    v := vcon.New("example.com")
    v.Subject = "Customer support call"

    // Add parties
    callerIdx := v.AddParty(vcon.Party{
        Name: "Alice Johnson",
        Tel:  "tel:+12025551234",
    })
    agentIdx := v.AddParty(vcon.Party{
        Name: "Bob Smith",
        Tel:  "tel:+18005559876",
    })

    // Add a dialog
    now := time.Now().UTC()
    v.AddDialog(vcon.Dialog{
        Type:       "recording",
        StartTime:  &now,
        Duration:   185.5,
        Parties:    []int{callerIdx, agentIdx},
        Originator: callerIdx,
        MediaType:  "audio/wav",
        URL:        "https://recordings.example.com/call-123.wav",
    })

    // Validate
    if err := v.Validate(); err != nil {
        fmt.Printf("Validation error: %v\n", err)
        return
    }

    fmt.Printf("Created vCon %s (v%s)\n", v.UUID, v.Vcon)
}
```

---

## Library Reference

### Creating a vCon

The `New()` function creates a vCon with a generated UUID v8, a timestamp, and empty slices for all collections:

```go
import "github.com/robjsliwa/go-vcon/pkg/vcon"

v := vcon.New("example.com")
v.Subject = "Weekly team standup"
```

Load from existing JSON:

```go
// From a JSON string
v, err := vcon.BuildFromJSON(jsonString)

// From a file
v, err := vcon.LoadFromFile("conversation.vcon.json")

// From a URL
v, err := vcon.LoadFromURL("https://api.example.com/vcons/123")
```

> **v0.0.3 Compatibility:** `BuildFromJSON` and `LoadFromFile` automatically detect
> v0.0.3 vCons and migrate them to v0.4.0 format. This includes converting `"base64"`
> encoding to `"base64url"`, removing deprecated fields (`alg`, `signature`, `appended`,
> `meta`), and reformatting content hashes.

### Parties

Parties represent conversation participants. Each party is identified by one or more
communication addresses:

```go
// Using struct literals
idx := v.AddParty(vcon.Party{
    Name:   "Alice Johnson",
    Tel:    "tel:+12025551234",
    Mailto: "mailto:alice@example.com",
})

// Using the option-based constructor
party := vcon.NewParty(
    vcon.WithName("Bob Smith"),
    vcon.WithTel("tel:+18005559876"),
    vcon.WithSip("sip:bob@pbx.example.com"),
)
idx = v.AddParty(*party)

// Find a party by property
index := v.FindPartyIndex("tel", "tel:+12025551234")
```

Supported address types: `Tel`, `Mailto`, `Sip`, `Did`, `Stir`.

### Dialogs

Dialogs represent individual conversation interactions -- calls, messages, transfers:

```go
now := time.Now().UTC()

// A recording dialog
v.AddDialog(vcon.Dialog{
    Type:       "recording",
    StartTime:  &now,
    Duration:   300.0,
    Parties:    []int{0, 1},
    Originator: 0,
    MediaType:  "audio/wav",
    Body:       "base64url-encoded-audio-data",
    Encoding:   "base64url",
})

// A text dialog
v.AddDialog(vcon.Dialog{
    Type:      "text",
    StartTime: &now,
    Parties:   []int{0, 1},
    Body:      "Hello, how can I help you today?",
    MediaType: "text/plain",
    Encoding:  "none",
})

// A transfer dialog with IntOrSlice fields
transferTime := now.Add(5 * time.Minute)
v.AddDialog(vcon.Dialog{
    Type:           "transfer",
    StartTime:      &transferTime,
    Transferee:     1,
    Transferor:     0,
    TransferTarget: vcon.NewIntValue(2),
    TargetDialog:   vcon.NewIntValue(0),
})
```

Dialog types: `"recording"`, `"text"`, `"transfer"`, `"incomplete"`.

Valid encodings: `"base64url"`, `"json"`, `"none"`.

#### External Data

Dialogs can reference externally hosted content instead of inlining it:

```go
dialog := &vcon.Dialog{
    Type:      "recording",
    StartTime: &now,
}
err := dialog.AddExternalData(
    "https://storage.example.com/call-123.wav",
    "call-123.wav",
    "audio/wav",
)
// This fetches the file, computes a SHA-512 content hash, and sets URL + ContentHash
```

#### Party History

Track participants joining, leaving, or being placed on hold during a dialog:

```go
v.AddDialog(vcon.Dialog{
    Type:      "recording",
    StartTime: &startTime,
    Duration:  900.0,
    Parties:   []int{0, 1, 2},
    PartyHistory: []vcon.PartyHistory{
        {Party: 1, Event: string(vcon.PartyEventJoin), Time: joinTime},
        {Party: 1, Event: string(vcon.PartyEventHold), Time: holdTime},
        {Party: 1, Event: string(vcon.PartyEventUnhold), Time: unholdTime},
        {Party: 2, Event: string(vcon.PartyEventJoin), Time: p2JoinTime},
        {Party: 1, Event: string(vcon.PartyEventDrop), Time: dropTime},
    },
})
```

Event types: `join`, `drop`, `hold`, `unhold`, `mute`, `unmute`, `keydown`, `keyup`.

### Analysis

Analysis entries hold derived data such as transcripts, sentiment scores, or speaker identification:

```go
v.AddAnalysis(vcon.Analysis{
    Type:      "transcript",
    Dialog:    []int{0},
    MediaType: "text/plain",
    Vendor:    "TranscriptCo",
    Product:   "AutoTranscribe v2.0",
    Body:      "Customer: Hi, I need help with my account...",
    Encoding:  "none",
})

v.AddAnalysis(vcon.Analysis{
    Type:      "sentiment",
    Dialog:    []int{0},
    MediaType: "application/json",
    Vendor:    "EmotionAI",
    Product:   "SentimentAnalyzer v3.1",
    Body:      `{"overall":"positive","customer":"satisfied","agent":"helpful"}`,
    Encoding:  "json",
})
```

### Attachments

Attachments are supplementary files associated with specific parties and time ranges:

```go
v.AddAttachment(vcon.Attachment{
    DialogIdx: vcon.IntPtr(0),
    PartyIdx:  1,
    StartTime: time.Now().UTC(),
    MediaType: "application/pdf",
    Filename:  "case_notes.pdf",
    Body:      "base64url-encoded-pdf-content",
    Encoding:  "base64url",
    Purpose:   "documentation",
})
```

### Validation

Validate a vCon against the JSON Schema and structural rules:

```go
// Returns an error with details
err := v.Validate()
if err != nil {
    fmt.Printf("Invalid: %v\n", err)
}

// Or get a bool and a list of specific issues
valid, errors := v.IsValid()
if !valid {
    for _, e := range errors {
        fmt.Println("  -", e)
    }
}
```

Validation checks include:
- JSON Schema compliance (draft-ietf-vcon-vcon-core-02)
- Valid party index references in dialogs and attachments
- Required fields (`uuid`, `created_at`, `parties`)
- Mutual exclusivity of `redacted`, `amended`, and `group`
- Critical extension support

### Signing and Verification

Sign a vCon using RS256 (JWS General JSON Serialization with detached payload):

```go
import (
    "crypto/rsa"
    "crypto/x509"
)

// Sign with a private key and certificate chain
signed, err := v.Sign(privateKey, []*x509.Certificate{cert})

// Verify against a trust anchor pool
rootPool := x509.NewCertPool()
rootPool.AddCert(caCert)

verified, err := signed.Verify(rootPool)
// verified is a *VCon with the original content
```

The signing process:
1. Serializes the vCon to [RFC 8785](https://datatracker.ietf.org/doc/rfc8785/) canonical JSON
2. Creates a JWS with `cty: application/vcon`, `x5c` certificate chain, and `uuid` header
3. Produces General JSON Serialization

### Encryption and Decryption

Encrypt a signed vCon for one or more recipients (JWE with RSA-OAEP + A256CBC-HS512):

```go
import "github.com/go-jose/go-jose/v4"

// Encrypt for a recipient
recipient := jose.Recipient{
    Algorithm: jose.RSA_OAEP,
    Key:       &recipientPublicKey,
}
encrypted, err := signed.Encrypt([]jose.Recipient{recipient})

// Decrypt with the recipient's private key
decrypted, err := encrypted.Decrypt(recipientPrivateKey)

// Convert back to SignedVCon for verification
signedVCon := vcon.SignedVCon{JSON: decrypted}
original, err := signedVCon.Verify(rootPool)
```

### Redaction

Create a redacted copy of a vCon while preserving structural indices (per Section 4.1.8):

```go
redacted, err := v.Redact("audio", func(copy *vcon.VCon) error {
    // Remove sensitive audio data but keep the dialog structure
    copy.Dialog[0].Body = ""
    copy.Dialog[0].Encoding = ""
    // Remove PII from parties
    copy.Parties[0].Tel = ""
    return nil
})

// redacted.Redacted.UUID points back to the original
// redacted.UUID is a new identifier for the redacted version
```

Optionally include a URL and hash pointing to the original:

```go
redacted, err := v.Redact("audio", redactFn,
    vcon.WithRedactedURL("https://archive.example.com/originals/123",
        vcon.ContentHashList{vcon.ComputeSHA512(originalData)}),
)
```

### Amendment

Create an amended copy with additional data (per Section 4.1.9):

```go
amended, err := v.Amend(func(copy *vcon.VCon) error {
    // Add a transcript that was generated after the original vCon
    copy.AddAnalysis(vcon.Analysis{
        Type:    "transcript",
        Dialog:  []int{0},
        Vendor:  "TranscriptCo",
        Product: "AutoTranscribe v2.0",
        Body:    "Full transcript text...",
        Encoding: "none",
    })
    return nil
})

// amended.Amended.UUID points back to the original
```

### Extensions

The extension framework allows adding custom parameters to vCon objects. Extensions are
registered in an `ExtensionRegistry` and declared in the vCon's `extensions` array.

#### Using the Contact Center (CC) Extension

The CC extension is auto-registered in `DefaultRegistry` via its `init()` function:

```go
import (
    "github.com/robjsliwa/go-vcon/pkg/vcon"
    _ "github.com/robjsliwa/go-vcon/pkg/vcon/ext/cc" // auto-registers
)

v := vcon.New("example.com")
v.Extensions = []string{"CC"} // declare CC extension usage
```

Access CC extension fields through typed helpers:

```go
import "github.com/robjsliwa/go-vcon/pkg/vcon/ext/cc"

// Set CC fields on a party
partyMap := v.Parties[0].ToMap()
cc.SetPartyData(partyMap, cc.PartyData{
    Role:        "agent",
    ContactList: "VIP",
})

// Read CC fields from a party
data := cc.GetPartyData(partyMap)
fmt.Println(data.Role) // "agent"

// Set CC fields on a dialog
dialogMap := v.Dialog[0].ToMap()
cc.SetDialogData(dialogMap, cc.DialogData{
    Campaign:        "summer_sale",
    InteractionType: "inbound",
    InteractionID:   "INT-12345",
    Skill:           "billing",
})
```

#### Creating a Custom Extension

```go
type MyExtension struct{}

func (e MyExtension) Name() string             { return "MYEXT" }
func (e MyExtension) IsCompatible() bool        { return true }
func (e MyExtension) PartyParams() []string     { return []string{"department"} }
func (e MyExtension) DialogParams() []string    { return nil }
func (e MyExtension) AnalysisParams() []string  { return nil }
func (e MyExtension) AttachmentParams() []string { return nil }
func (e MyExtension) VConParams() []string       { return nil }

// Register it
vcon.DefaultRegistry.Register(MyExtension{})
```

#### Critical Extensions

Extensions listed in the `critical` array must be understood by any processor. If
a critical extension is not registered, validation will fail:

```go
v.Extensions = []string{"CC", "CUSTOM"}
v.Critical = []string{"CUSTOM"}  // processors MUST understand CUSTOM

err := v.Validate()
// Fails if "CUSTOM" is not registered in the extension registry
```

### Content Hashing

Content hashes use the format `"algorithm-base64url_encoded_hash"` and default to SHA-512:

```go
// Compute a SHA-512 hash
hash := vcon.ComputeSHA512(fileData)
fmt.Println(hash.String()) // "sha512-abc123..."

// Verify a hash
if hash.Verify(fileData) {
    fmt.Println("Content integrity verified")
}

// Parse an existing hash string
parsed, err := vcon.ParseContentHash("sha512-abc123...")

// ContentHashList handles JSON serialization:
// - single hash serializes as a string
// - multiple hashes serialize as an array
```

### Form Detection

Determine whether raw JSON is an unsigned vCon, a signed JWS, or an encrypted JWE:

```go
data, _ := os.ReadFile("conversation.json")

form, err := vcon.DetectForm(data)
switch form {
case vcon.VConFormUnsigned:
    fmt.Println("Unsigned vCon")
case vcon.VConFormSigned:
    fmt.Println("Signed vCon (JWS)")
case vcon.VConFormEncrypted:
    fmt.Println("Encrypted vCon (JWE)")
}
```

### Serialization

```go
// To JSON string
jsonStr := v.ToJSON()

// To map
m := v.ToMap()

// Save to file
err := v.SaveToFile("output.vcon.json")
```

---

## CLI Reference

```
vconctl - a tool for working with vCon files

Usage:
  vconctl [command]

Available Commands:
  convert     Convert external artifacts (audio, zoom, email) into vCon containers
  decrypt     Decrypt an encrypted vCon file
  detect      Detect the form of a vCon file (unsigned, signed, or encrypted)
  encrypt     Encrypt a signed vCon for one recipient
  genkey      Generate a test RSA key pair and self-signed certificate
  sign        Sign a vCon file using a private key and certificate
  validate    Validate a vCon file
  verify      Verify the signature on a signed vCon

Global Flags:
  --domain string   Domain name for UUID generation (default "vcon.example.com")
```

### validate

Validate one or more vCon files against the JSON Schema:

```bash
vconctl validate conversation.vcon.json

# Validate multiple files
vconctl validate file1.json file2.json file3.json
```

Output:

```
Validating conversation.vcon.json...
✅ conversation.vcon.json is valid
```

### detect

Identify the form of a vCon file:

```bash
vconctl detect conversation.json
```

Output:

```
conversation.json: unsigned
```

Possible results: `unsigned`, `signed`, `encrypted`, `unknown`.

### genkey

Generate a test RSA key pair and self-signed certificate:

```bash
# Default paths (test_key.pem, test_cert.pem)
vconctl genkey

# Custom paths
vconctl genkey --key my_key.pem --cert my_cert.pem
```

| Flag | Default | Description |
|------|---------|-------------|
| `--key, -k` | `test_key.pem` | Output private key path |
| `--cert, -c` | `test_cert.pem` | Output certificate path |

### sign

Sign a vCon file using RS256 with a private key and certificate:

```bash
vconctl sign conversation.vcon.json --key private.pem --cert certificate.pem

# Custom output path
vconctl sign conversation.vcon.json --key private.pem --cert certificate.pem -o signed.json
```

| Flag | Default | Description |
|------|---------|-------------|
| `--key, -k` | _(required)_ | Path to RSA private key (PEM) |
| `--cert, -c` | _(required)_ | Path to X.509 certificate (PEM) |
| `--output, -o` | `<file>.signed.json` | Output file path |

### verify

Verify the signature on a signed vCon:

```bash
vconctl verify conversation.signed.json --cert certificate.pem
```

| Flag | Default | Description |
|------|---------|-------------|
| `--cert, -c` | _(required)_ | Path to trust anchor certificate (PEM) |

### encrypt

Encrypt a signed vCon for a recipient:

```bash
vconctl encrypt conversation.signed.json --cert recipient_cert.pem

# Custom output path
vconctl encrypt conversation.signed.json --cert recipient_cert.pem -o encrypted.json
```

| Flag | Default | Description |
|------|---------|-------------|
| `--cert, -c` | _(required)_ | Path to recipient certificate (PEM) |
| `--output, -o` | `<file>.encrypted.json` | Output file path |

### decrypt

Decrypt an encrypted vCon:

```bash
vconctl decrypt conversation.encrypted.json --key private.pem

# Custom output path
vconctl decrypt conversation.encrypted.json --key private.pem -o decrypted.json
```

| Flag | Default | Description |
|------|---------|-------------|
| `--key, -k` | _(required)_ | Path to RSA private key (PEM) |
| `--output, -o` | `<file>.decrypted.json` | Output file path |

### convert audio

Create a vCon from a standalone audio recording. Requires `ffprobe` to be installed.

```bash
vconctl convert audio \
  --input recording.wav \
  --party "Alice,tel:+12025551234" \
  --party "Bob,tel:+12025555678" \
  --date "2025-07-20T14:30:00Z" \
  --domain example.com \
  -o call.vcon.json
```

Supports local files and remote URLs:

```bash
vconctl convert audio \
  --input https://example.com/recordings/call.wav \
  --party "Customer,tel:+12025551111" \
  --party "Agent,mailto:agent@example.com"
```

| Flag | Default | Description |
|------|---------|-------------|
| `--input` | _(required)_ | Path or URL to audio file |
| `--party` | _(repeatable)_ | Party spec: `name,tel:+1...` or `name,mailto:...` or `name,sip:...` or `name,did:...` |
| `--date` | file mtime | Recording start time (RFC 3339) |
| `--output, -o` | `<input>.vcon.json` | Output file path |
| `--domain` | `vcon.example.com` | Domain for UUID generation |

### convert zoom

Create a vCon from a Zoom recording folder:

```bash
vconctl convert zoom ./zoom_meeting_folder
```

The command reads metadata from `meeting_info.json` or `recording.conf` and enumerates
media files (`.mp4`, `.m4a`, `.mov`, `.vtt`, `.txt`). Host and participant information
is extracted from the metadata.

### convert email

Create a vCon from an RFC-822 email message:

```bash
vconctl convert email message.eml -o email.vcon.json
```

Parses `From`, `To`, `Cc`, `Subject`, `Date`, and `Message-ID` headers. The email body
becomes a text dialog.

| Flag | Default | Description |
|------|---------|-------------|
| `--output, -o` | `<file>.vcon.json` | Output file path |

---

## Complete Workflow Examples

### Validate, Sign, Encrypt, Decrypt, Verify

```bash
# 1. Generate test keys
vconctl genkey

# 2. Validate the original vCon
vconctl validate conversation.vcon.json

# 3. Sign it
vconctl sign conversation.vcon.json --key test_key.pem --cert test_cert.pem
# -> conversation.vcon.signed.json

# 4. Verify the signature
vconctl verify conversation.vcon.signed.json --cert test_cert.pem

# 5. Encrypt the signed vCon
vconctl encrypt conversation.vcon.signed.json --cert test_cert.pem
# -> conversation.vcon.signed.encrypted.json

# 6. Decrypt it
vconctl decrypt conversation.vcon.signed.encrypted.json --key test_key.pem
# -> conversation.vcon.signed.encrypted.decrypted.json

# 7. Verify the decrypted content
vconctl verify conversation.vcon.signed.encrypted.decrypted.json --cert test_cert.pem
```

### Detect and Process

```bash
# Detect the form first, then process accordingly
vconctl detect unknown_file.json

# If signed:
vconctl verify unknown_file.json --cert ca.pem

# If encrypted:
vconctl decrypt unknown_file.json --key private.pem
```

### Convert and Validate

```bash
# Convert a recording to vCon
vconctl convert audio \
  --input meeting.wav \
  --party "John Doe,tel:+12025551111" \
  --party "Jane Smith,tel:+12025552222" \
  --date "2025-07-20T14:30:00Z" \
  --domain mycompany.com \
  -o meeting.vcon.json

# Validate the result
vconctl validate meeting.vcon.json

# Sign it for archival
vconctl sign meeting.vcon.json --key production_key.pem --cert production_cert.pem
```

### Full Library Workflow

```go
package main

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "fmt"
    "time"

    "github.com/go-jose/go-jose/v4"
    "github.com/robjsliwa/go-vcon/pkg/vcon"
    _ "github.com/robjsliwa/go-vcon/pkg/vcon/ext/cc"
)

func main() {
    // Create a vCon
    v := vcon.New("example.com")
    v.Subject = "Support call #4521"
    v.Extensions = []string{"CC"}

    callerIdx := v.AddParty(vcon.Party{
        Name: "Alice Johnson",
        Tel:  "tel:+12025551234",
    })
    agentIdx := v.AddParty(vcon.Party{
        Name: "Bob Smith",
        Tel:  "tel:+18005559876",
    })

    now := time.Now().UTC()
    dialogIdx := v.AddDialog(vcon.Dialog{
        Type:       "recording",
        StartTime:  &now,
        Duration:   185.5,
        Parties:    []int{callerIdx, agentIdx},
        Originator: callerIdx,
        MediaType:  "audio/wav",
    })

    v.AddAnalysis(vcon.Analysis{
        Type:      "transcript",
        Dialog:    []int{dialogIdx},
        MediaType: "text/plain",
        Body:      "Alice: Hi, I need help...\nBob: Of course...",
        Encoding:  "none",
    })

    // Validate
    if err := v.Validate(); err != nil {
        panic(err)
    }

    // Sign (using test keys for demonstration)
    privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
    // In production, load your key and certificate from files
    signed, err := v.Sign(privateKey, nil)
    if err != nil {
        panic(err)
    }

    // Encrypt
    recipient := jose.Recipient{
        Algorithm: jose.RSA_OAEP,
        Key:       &privateKey.PublicKey,
    }
    encrypted, err := signed.Encrypt([]jose.Recipient{recipient})
    if err != nil {
        panic(err)
    }

    // Decrypt
    decrypted, err := encrypted.Decrypt(privateKey)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Round-trip successful. UUID: %s\n", decrypted["uuid"])

    // Save
    v.SaveToFile("conversation.vcon.json")
}
```

---

## Sample vCon Files

### Simple vCon

Save as `simple-vcon.json`:

```json
{
  "vcon": "0.4.0",
  "uuid": "9b583dd6-31b2-4403-b74e-271f45f97ada",
  "created_at": "2025-06-15T14:25:33Z",
  "subject": "Customer Support Call",
  "parties": [
    {
      "name": "John Doe",
      "tel": "+12025551234"
    },
    {
      "name": "Jane Smith",
      "tel": "+18005559876"
    }
  ]
}
```

```bash
$ vconctl validate simple-vcon.json
Validating simple-vcon.json...
✅ simple-vcon.json is valid
```

### vCon with Dialogs and Analysis

Save as `full-vcon.json`:

```json
{
  "vcon": "0.4.0",
  "uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "created_at": "2025-06-15T14:30:00Z",
  "subject": "Technical Support - Network Issue",
  "parties": [
    {
      "name": "Bob Johnson",
      "tel": "+12025551111"
    },
    {
      "name": "Sarah Lee",
      "tel": "+18005552222"
    }
  ],
  "dialog": [
    {
      "type": "text",
      "start": "2025-06-15T14:30:00Z",
      "duration": 300,
      "parties": [0, 1],
      "originator": 0,
      "mediatype": "text/plain",
      "body": "Customer reports intermittent network drops since firmware update.",
      "encoding": "none"
    }
  ],
  "analysis": [
    {
      "type": "sentiment",
      "vendor": "EmotionAI",
      "product": "SentimentAnalyzer v3.1",
      "body": "{\"customer\": \"frustrated\", \"agent\": \"helpful\"}",
      "encoding": "json"
    }
  ],
  "attachments": [
    {
      "body": "bmV0d29ya19sb2dzX2hlcmU",
      "encoding": "base64url",
      "party": 0,
      "start": "2025-06-15T14:30:00Z",
      "purpose": "diagnostics"
    }
  ]
}
```

### Legacy v0.0.3 vCon (Auto-Migrated)

v0.0.3 vCons are automatically migrated when loaded:

```json
{
  "vcon": "0.0.3",
  "uuid": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "created_at": "2023-06-15T14:30:00Z",
  "parties": [
    { "name": "Alice", "tel": "+12025551234", "role": "customer" }
  ],
  "dialog": [
    {
      "type": "text",
      "start": "2023-06-15T14:30:00Z",
      "parties": [0],
      "body": "Hello",
      "encoding": "base64",
      "alg": "sha256",
      "signature": "abc123"
    }
  ]
}
```

When loaded with `BuildFromJSON` or `LoadFromFile`:
- `vcon` is updated to `"0.4.0"`
- `encoding` `"base64"` becomes `"base64url"`
- `alg` and `signature` fields are removed
- `role` and other CC extension fields are removed from core objects
- `content_hash` separator `:` is converted to `-`

---

## Development

### Running Tests

```bash
# All tests
go test ./...

# Library tests only
go test ./pkg/vcon/...

# CC extension tests
go test ./pkg/vcon/ext/cc/...

# CLI tests
go test ./cmd/vconctl/...

# A single test
go test -run TestSignAndVerify ./pkg/vcon/...

# Verbose output
go test -v ./pkg/vcon/...
```

### Test Coverage

```bash
# Summary
go test -cover ./...

# HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Function-level breakdown
go tool cover -func=coverage.out
```

### Static Analysis

```bash
go vet ./...
```

### Project Structure

```
go-vcon/
├── cmd/vconctl/          # CLI tool
│   ├── main.go           # Root command, flags, helpers
│   ├── validate.go       # validate command
│   ├── sign.go           # sign command
│   ├── keys.go           # genkey + verify commands
│   ├── encrypt.go        # encrypt + decrypt commands
│   ├── detect.go         # detect command
│   ├── convert_audio.go  # convert audio
│   ├── convert_zoom.go   # convert zoom
│   └── convert_email.go  # convert email
├── pkg/vcon/             # Core library
│   ├── vcon.go           # VCon type, constructors, validation
│   ├── party.go          # Party type
│   ├── dialog.go         # Dialog type, MIME types
│   ├── attachment.go     # Attachment type
│   ├── content_hash.go   # SHA-512 content hashing
│   ├── types.go          # RedactedObject, AmendedObject, IntOrSlice
│   ├── extension.go      # Extension interface and registry
│   ├── crypto.go         # JWS/JWE signing and encryption
│   ├── canonical.go      # RFC 8785 canonicalization
│   ├── civ_address.go    # Civic address (RFC 5139)
│   ├── form.go           # Form detection
│   ├── compress.go       # Gzip compression
│   ├── redact.go         # Redaction workflow
│   ├── amend.go          # Amendment workflow
│   ├── schema/
│   │   └── vcon.json     # Embedded JSON Schema
│   └── ext/cc/
│       └── cc.go         # Contact Center extension
└── testdata/             # Test fixtures
    └── sample_vcons/     # Sample vCon files, keys, audio
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run the tests (`go test ./...`)
4. Commit your changes (`git commit -m 'Add some amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---

## Acknowledgments

- [IETF vCon Working Group](https://datatracker.ietf.org/wg/vcon/about/)
- [draft-ietf-vcon-vcon-core-02](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-core/) -- vCon core specification
- [draft-ietf-vcon-cc-extension-01](https://datatracker.ietf.org/doc/draft-ietf-vcon-cc-extension/) -- Contact Center extension
