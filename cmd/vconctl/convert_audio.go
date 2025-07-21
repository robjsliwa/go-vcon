package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
	"github.com/vansante/go-ffprobe"
)

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

	var dialogParties []int

	for _, spec := range audioParties {
		p := parseParty(spec)
		v.Parties = append(v.Parties, *p)
		dialogParties = append(dialogParties, len(v.Parties)-1)
	}

	dur := time.Duration(float64(time.Second) * info.Format.DurationSeconds)
	v.Dialog = append(v.Dialog, vcon.Dialog{
		Type:      "recording",
		StartTime: &v.CreatedAt,
		Duration:  dur.Seconds(),
		Parties:   dialogParties,
		Filename:  filepath.Base(path),
		MediaType: strings.ReplaceAll(info.Format.FormatName, ",", "/"),
		URL:       audioInput,
	})

	return writeVconFile(v, vConOut, path)
}