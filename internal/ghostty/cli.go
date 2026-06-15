// Package ghostty is a thin wrapper over the optional `ghostty` binary, used to
// validate configs and discover themes/fonts/options. All methods degrade
// gracefully when the binary is not available.
package ghostty

import (
	"bytes"
	"os/exec"
	"strings"
)

// CLI represents a located ghostty binary. Path is empty if none was found.
type CLI struct {
	Path string
}

// Detect locates the ghostty binary on PATH.
func Detect() *CLI {
	path, err := exec.LookPath("ghostty")
	if err != nil {
		return &CLI{}
	}
	return &CLI{Path: path}
}

// Available reports whether a ghostty binary was found.
func (c *CLI) Available() bool { return c.Path != "" }

// ValidateConfig runs `ghostty +validate-config --config-file=<path>` and
// reports whether it succeeded along with combined output.
func (c *CLI) ValidateConfig(path string) (ok bool, output string, err error) {
	if !c.Available() {
		return false, "", errUnavailable
	}
	out, runErr := c.run("+validate-config", "--config-file="+path)
	return runErr == nil, out, nil
}

// ListThemes returns available theme names.
func (c *CLI) ListThemes() ([]string, error) {
	out, err := c.run("+list-themes", "--plain")
	if err != nil {
		return nil, err
	}
	return parseList(out), nil
}

// ListFonts returns available font family names.
func (c *CLI) ListFonts() ([]string, error) {
	out, err := c.run("+list-fonts")
	if err != nil {
		return nil, err
	}
	return parseList(out), nil
}

// ShowConfigDocs returns the output of `+show-config --default --docs`.
func (c *CLI) ShowConfigDocs() (string, error) {
	return c.run("+show-config", "--default", "--docs")
}

func (c *CLI) run(args ...string) (string, error) {
	if !c.Available() {
		return "", errUnavailable
	}
	var buf bytes.Buffer
	cmd := exec.Command(c.Path, args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// parseList splits command output into trimmed, non-empty lines.
func parseList(out string) []string {
	var res []string
	for _, line := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			res = append(res, t)
		}
	}
	return res
}

type cliError string

func (e cliError) Error() string { return string(e) }

const errUnavailable = cliError("ghostty binary not available")
