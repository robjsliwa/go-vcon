package main

import (
	"fmt"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/spf13/cobra"
)

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