// Command smg is a GUI editor for ghostty configuration files.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := flag.String("config", "", "path to ghostty config file (default: auto-detect)")
	flag.Parse()

	// UI wiring is added in a later task; for now just prove the binary builds.
	if *configPath != "" {
		fmt.Fprintf(os.Stdout, "smg: using config %s\n", *configPath)
	}
}
