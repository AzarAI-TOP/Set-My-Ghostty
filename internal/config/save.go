package config

import (
	"os"
	"path/filepath"
)

// Save writes a .bak backup of the current file (if it exists) and then writes
// the document. If the backup cannot be written, the save is aborted and the
// original file is left untouched.
func Save(path string, d *Document) error {
	if existing, err := os.ReadFile(path); err == nil {
		if err := os.WriteFile(path+".bak", existing, 0o644); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, d.Bytes(), 0o644)
}

// ResolvePath picks the config file to edit. Order:
//  1. flagPath, if non-empty.
//  2. $XDG_CONFIG_HOME/ghostty/config (XDG_CONFIG_HOME defaults to ~/.config).
//  3. ~/.config/ghostty/config.ghostty (this user's existing file).
//  4. The canonical default path even if it does not yet exist.
func ResolvePath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	canonical := filepath.Join(base, "ghostty", "config")
	candidates := []string{
		canonical,
		filepath.Join(base, "ghostty", "config.ghostty"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return canonical, nil
}
