package main

import (
	"fmt"
	"os"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

var detectCmd = &cobra.Command{
	Use:   "detect <file>",
	Short: "Detect the form of a vCon file (unsigned, signed, or encrypted)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDetect,
}

func runDetect(_ *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	form, err := vcon.DetectForm(data)
	if err != nil {
		return fmt.Errorf("detect form: %w", err)
	}

	fmt.Printf("%s: %s\n", args[0], form)
	return nil
}
