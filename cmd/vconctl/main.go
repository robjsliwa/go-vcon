package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
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
	vConOut     string
	
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
	audioCmd.Flags().StringVarP(&vConOut,   "output", "o", "",  "Output vCon (default: <rec>.json)")
	audioCmd.MarkFlagRequired("input")

	emailCmd.Flags().StringVarP(&vConOut,   "output", "o", "",  "Output vCon (default: <file>.json)")
}

func die(context string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s: %v\n", context, err)
	os.Exit(1)
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
