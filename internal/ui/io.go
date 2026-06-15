package ui

import (
	"io"
	"os"
	"strings"
)

func readFileAllowMissing(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return b, err
}

func stringsReader(s string) io.Reader { return strings.NewReader(s) }
