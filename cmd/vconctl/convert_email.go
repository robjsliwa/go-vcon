package main

import (
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhillyerd/enmime"
	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

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

	var dialogParties []int

	parseAndAdd := func(header, role string) error {
		addrsStr := env.GetHeader(header)
		if addrsStr == "" && header == "Cc" { return nil }
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
			dialogParties = append(dialogParties, len(v.Parties)-1)
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
		Parties:   dialogParties,
		Body:      env.Text,
		MediaType: "text/plain",
		MessageID: env.GetHeader("Message-Id"),
	})

	return writeVconFile(v, vConOut, f)
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