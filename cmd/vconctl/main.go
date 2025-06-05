package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: vconctl validate <file.json>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "validate":
		runValidate(os.Args[2:])
	default:
		fmt.Printf("unknown command %q\n", os.Args[1])
	}
}

func runValidate(args []string) {
	for _, path := range args {
		b, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("%s: %v\n", path, err)
			continue
		}
		var v vcon.VCon
		if err := json.Unmarshal(b, &v); err != nil {
			fmt.Printf("%s: %v\n", path, err)
			continue
		}
		if err := v.Validate(); err != nil {
			fmt.Printf("%s: invalid: %v\n", path, err)
		} else {
			fmt.Printf("%s: OK\n", path)
		}
	}
}
