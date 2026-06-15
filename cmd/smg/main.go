// Command smg is a GUI editor for ghostty configuration files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/ui"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	configPath := flag.String("config", "", "path to ghostty config file (default: auto-detect)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("smg", version)
		return
	}

	path, err := config.ResolvePath(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "smg:", err)
		os.Exit(1)
	}
	if err := ui.Run(path); err != nil {
		fmt.Fprintln(os.Stderr, "smg:", err)
		os.Exit(1)
	}
}
