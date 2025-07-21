package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/jhillyerd/enmime"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
	ffprobe "github.com/vansante/go-ffprobe"
)

var rootCmd = &cobra.Command{
	Use:   "vconctl",
	Short: "vconctl - a tool for working with vCon files",
	Long:  `vconctl is a command-line utility for validating, signing, encrypting, verifying, and decrypting vCon (Virtual Conversation) files.`,
}

var (
	audioInput  string
	audioParties []string
	audioDate   string
	audioOut    string
	
	// Global domain flag for UUID generation
	globalDomain string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert external artefacts (audio, Zoom, email) into vCon containers",
}

func main() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(validateCmd, signCmd, encryptCmd, verifyCmd, decryptCmd, genkeyCmd, convertCmd)
	convertCmd.AddCommand(audioCmd, zoomCmd, emailCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&globalDomain, "domain", "vcon.example.com", "Domain name for UUID generation")

	// flags
	signCmd.Flags().StringP("key", "k", "", "Path to private key file (required)")
	signCmd.Flags().StringP("cert", "c", "", "Path to certificate file (required)")
	signCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to <file>.signed.json)")

	encryptCmd.Flags().StringP("cert", "c", "", "Path to recipient certificate (required)")
	encryptCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to <file>.encrypted.json)")

	verifyCmd.Flags().StringP("cert", "c", "", "Path to trust anchor (leaf or CA) (required)")

	decryptCmd.Flags().StringP("key", "k", "", "Path to private key file (required)")
	decryptCmd.Flags().StringP("output", "o", "", "Path to output file (defaults to <file>.decrypted.json)")

	genkeyCmd.Flags().StringP("key", "k", "", "Output private-key path (default: test_key.pem)")
	genkeyCmd.Flags().StringP("cert", "c", "", "Output certificate path (default: test_cert.pem)")

	audioCmd.Flags().StringVar(&audioInput,  "input",  "",  "Path or URL to recording (required)")
	audioCmd.Flags().StringArrayVar(&audioParties, "party", nil, "Party spec 'name,tel:+1555...' or 'name,mailto:bob@a.b'")
	audioCmd.Flags().StringVar(&audioDate,   "date",   "",  "Recording start (RFC3339); default file mtime")
	audioCmd.Flags().StringVarP(&audioOut,   "output", "o", "",  "Output vCon (default: <rec>.json)")
	audioCmd.MarkFlagRequired("input")
}

// Command: validate

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a vCon file",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		for _, p := range args {
			fmt.Printf("Validating %s…\n", p)
			if _, err := vcon.LoadFromFile(p, vcon.PropertyHandlingStrict); err != nil {
				fmt.Printf("❌ %v\n", err)
				continue
			}
			fmt.Printf("✅ %s is valid\n", p)
		}
	},
}

// Command: sign

var signCmd = &cobra.Command{
	Use:   "sign [file]",
	Short: "Sign a vCon file using a private key and certificate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")
		certPath, _ := cmd.Flags().GetString("cert")
		outPath, _ := cmd.Flags().GetString("output")
		if keyPath == "" || certPath == "" {
			fmt.Println("Error: --key and --cert are required")
			_ = cmd.Help()
			os.Exit(1)
		}
		signFile(args[0], keyPath, certPath, outPath)
	},
}

func signFile(path, keyPath, certPath, outPath string) {
	fmt.Printf("Signing %s…\n", path)

	raw, err := os.ReadFile(path)
	if err != nil {
		die("reading vCon", err)
	}
	var v vcon.VCon
	if err := json.Unmarshal(raw, &v); err != nil {
		die("parsing JSON", err)
	}

	priv := readPrivateKey(keyPath)
	cert := readCertificate(certPath)

	signed, err := v.Sign(priv, []*x509.Certificate{cert})
	if err != nil {
		die("signing vCon", err)
	}

	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".signed" + ext
	}
	if err := writeJSON(outPath, signed.JSON); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Signed vCon written to %s\n", outPath)
}

// Command: verify

var verifyCmd = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verify the signature on a signed vCon",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		caPath, _ := cmd.Flags().GetString("cert")
		if caPath == "" {
			fmt.Println("Error: --cert is required")
			_ = cmd.Help()
			os.Exit(1)
		}
		verifyFile(args[0], caPath)
	},
}

func verifyFile(path, caPath string) {
	fmt.Printf("Verifying %s…\n", path)

	jwsMap := readBareJWS(path)

	root := x509.NewCertPool()
	if ok := appendPEMToPool(root, caPath); !ok {
		die("loading trust anchor", fmt.Errorf("invalid PEM in %s", caPath))
	}

	signed := vcon.SignedVCon{JSON: jwsMap}
	vc, err := signed.Verify(root)
	if err != nil {
		die("signature verification failed", err)
	}

	fmt.Println("✅ Signature verified!")
	fmt.Printf("Subject : %s\nUUID    : %s\nCreated : %s\nParties : %d\n",
		vc.Subject, vc.UUID, vc.CreatedAt, len(vc.Parties))
}

// Command: encrypt

var encryptCmd = &cobra.Command{
	Use:   "encrypt [file]",
	Short: "Encrypt a signed vCon for one recipient",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		certPath, _ := cmd.Flags().GetString("cert")
		outPath, _ := cmd.Flags().GetString("output")
		if certPath == "" {
			fmt.Println("Error: --cert is required")
			_ = cmd.Help()
			os.Exit(1)
		}
		encryptFile(args[0], certPath, outPath)
	},
}

func encryptFile(path, certPath, outPath string) {
	fmt.Printf("Encrypting %s…\n", path)

	jwsMap := readBareJWS(path)
	signed := vcon.SignedVCon{JSON: jwsMap}
	cert := readCertificate(certPath)

	obj, err := signed.Encrypt([]jose.Recipient{{
		Algorithm: jose.RSA_OAEP,
		Key:       cert.PublicKey,
	}})
	if err != nil {
		die("encrypting", err)
	}

	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".encrypted" + ext
	}
	if err := writeJSON(outPath, obj); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Encrypted vCon written to %s\n", outPath)
}

// Command decrypt

var decryptCmd = &cobra.Command{
	Use:   "decrypt [file]",
	Short: "Decrypt an encrypted vCon file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyPath, _ := cmd.Flags().GetString("key")
		outPath, _ := cmd.Flags().GetString("output")
		if keyPath == "" {
			fmt.Println("Error: --key is required")
			_ = cmd.Help()
			os.Exit(1)
		}
		decryptFile(args[0], keyPath, outPath)
	},
}

func decryptFile(path, keyPath, outPath string) {
	fmt.Printf("Decrypting %s…\n", path)

	// Read encrypted JWE
	raw, err := os.ReadFile(path)
	if err != nil {
		die("reading file", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		die("parsing JSON", err)
	}

	jweContent, ok := m["jwe"]
	if !ok {
		die("extracting JWE", fmt.Errorf("no 'jwe' field found"))
	}
	jweMap, ok := jweContent.(map[string]any)
	if !ok {
		die("extracting JWE", fmt.Errorf("'jwe' field is not an object"))
	}

	encrypted := vcon.EncryptedVCon{JSON: jweMap}
	priv := readPrivateKey(keyPath)

	decrypted, err := encrypted.Decrypt(priv)
	if err != nil {
		die("decrypting", err)
	}

	if outPath == "" {
		ext := filepath.Ext(path)
		outPath = path[:len(path)-len(ext)] + ".decrypted" + ext
	}
	if err := writeJSON(outPath, decrypted); err != nil {
		die("writing output", err)
	}
	fmt.Printf("✅ Decrypted vCon written to %s\n", outPath)
}

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

func die(context string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s: %v\n", context, err)
	os.Exit(1)
}

// Command: audio

var audioCmd = &cobra.Command{
	Use:   "audio --input <file|url> --party <spec> [--party <spec> ...] --date <RFC3339>",
	Short: "Create a vCon from a standalone recording",
	Args:  cobra.NoArgs,
	RunE:  runAudio,
}

func runAudio(cmd *cobra.Command, _ []string) error {
	path, cleanup, err := fetchIfRemote(audioInput)
	if err != nil { return err }
	defer cleanup()

	info, err := ffprobe.GetProbeData(path, 10*time.Second)
	if err != nil { return fmt.Errorf("ffprobe: %w", err) }

	v := vcon.New(globalDomain)
	v.Subject   = filepath.Base(path)
	v.CreatedAt = getDate(audioDate, path)

	for _, spec := range audioParties {
		p := parseParty(spec)
		v.Parties = append(v.Parties, *p)
	}

	dur := time.Duration(float64(time.Second) * info.Format.DurationSeconds)
	v.Dialog = append(v.Dialog, vcon.Dialog{
		Type:      "recording",
		StartTime: &v.CreatedAt,
		Duration:  dur.Seconds(),
		Filename:  filepath.Base(path),
		MediaType: strings.ReplaceAll(info.Format.FormatName, ",", "/"),
		URL:       audioInput,
	})

	return writeVconFile(v, audioOut, path)
}

// Command: zoom
var zoomCmd = &cobra.Command{
	Use:   "zoom <folder>",
	Short: "Generate a vCon from a local Zoom recording folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runZoom,
}

func runZoom(_ *cobra.Command, args []string) error {
	folder := args[0]
	meta, err := readZoomMeta(folder)
	if err != nil { return err }

	v := vcon.New(globalDomain)
	v.Subject   = meta.Topic
	v.CreatedAt = meta.Start

	// host
	v.Parties = append(v.Parties, vcon.Party{ Name: meta.Host, Mailto: meta.HostEmail, Role: "host" })
	// participants
	for _, p := range meta.Participants {
		v.Parties = append(v.Parties, vcon.Party{ Name: p.Name, Mailto: p.Email })
	}

	// main MP4 and VTT transcript become attachments
	for _, f := range meta.Files {
		att := vcon.Attachment{
			Filename: f.Name,
			URL:      f.Path,
			MediaType: f.Type,
		}
		v.Attachments = append(v.Attachments, att)
	}

	return writeVconFile(v, "", folder)
}

// Command: email
var emailCmd = &cobra.Command{
	Use:   "email <file.eml>",
	Short: "Convert a raw RFC-822 mail into vCon",
	Args:  cobra.ExactArgs(1),
	RunE:  runEmail,
}

func runEmail(_ *cobra.Command, args []string) error {
	f := args[0]
	r, err := os.Open(f); if err != nil { return err }
	defer r.Close()

	env, err := enmime.ReadEnvelope(r)
	if err != nil { return err }

	v := vcon.New(globalDomain)
	v.Subject   = env.GetHeader("Subject")
	dateStr := env.GetHeader("Date")
	created, err := mail.ParseDate(dateStr)
	if err != nil {
		return fmt.Errorf("parsing Date header: %w", err)
	}
	v.CreatedAt = created

	parseAndAdd := func(header, role string) error {
		addrsStr := env.GetHeader(header)
		addrs, err := mail.ParseAddressList(addrsStr)
		if err != nil {
			return fmt.Errorf("parsing %s header: %w", header, err)
		}
		for _, a := range addrs {
			v.Parties = append(v.Parties, vcon.Party{
				Name:   a.Name,
				Mailto: "mailto:" + a.Address,
				Role:   role,
			})
		}
		return nil
	}

	if err := parseAndAdd("From", "originator"); err != nil {
		return err
	}
	if err := parseAndAdd("To", "recipient"); err != nil {
		return err
	}
	if err := parseAndAdd("Cc", "cc"); err != nil {
		return err
	}

	v.Dialog = append(v.Dialog, vcon.Dialog{
		Type:      "email",
		StartTime: &v.CreatedAt,
		Parties:   []int{0}, // index 0 assumed as originator
		Body:      env.Text,
		MediaType: "text/plain",
		MessageID: env.GetHeader("Message-Id"),
	})

	return writeVconFile(v, "", f)
}

// helpers
func fetchIfRemote(src string) (path string, cleanup func(), err error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		downloadURL := src
		
		tmp, err := os.CreateTemp("", "vcon-dl-*"+filepath.Ext(src))
		if err != nil { return "", nil, err }
		
		resp, err := http.Get(downloadURL)
		if err != nil { return "", nil, err }
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return "", nil, fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
		}
		
		_, err = io.Copy(tmp, resp.Body)
		if err != nil { return "", nil, err }
		tmp.Close()
		
		// Verify the file was downloaded correctly by checking its size
		if stat, err := os.Stat(tmp.Name()); err == nil {
			fmt.Printf("Downloaded %d bytes to %s\n", stat.Size(), tmp.Name())
		}
		
		return tmp.Name(), func(){ os.Remove(tmp.Name()) }, nil
	}
	return src, func(){}, nil
}

func parseParty(spec string) *vcon.Party {
	parts := strings.SplitN(spec, ",", 2)
	p := &vcon.Party{ Name: parts[0] }
	if len(parts) == 2 {
		if strings.HasPrefix(parts[1], "tel:") { p.Tel = parts[1] }
		if strings.HasPrefix(parts[1], "mailto:") { p.Mailto = parts[1] }
	}
	return p
}

func getDate(flag, path string) time.Time {
	if flag != "" {
		if t, err := time.Parse(time.RFC3339, flag); err == nil {
			return t
		}
	}
	if fi, err := os.Stat(path); err == nil {
		return fi.ModTime()
	}
	return time.Now()
}

func writeVconFile(v *vcon.VCon, out, src string) error {
	if out == "" {
		out = strings.TrimSuffix(src, filepath.Ext(src)) + ".vcon.json"
	}
	blob, _ := json.MarshalIndent(v, "", "  ")
	return os.WriteFile(out, blob, 0644)
}
