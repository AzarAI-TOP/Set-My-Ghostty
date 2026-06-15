package schema

import (
	"bufio"
	"io"
	"strings"
)

// ParseShowConfigDocs parses the output of `ghostty +show-config --default
// --docs` into a Schema. Doc-comment lines (starting with '#') accumulate until
// the next `key = value` line, which they describe.
func ParseShowConfigDocs(r io.Reader) (*Schema, error) {
	s := &Schema{Options: map[string]Option{}}
	var docLines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			docLines = nil
		case strings.HasPrefix(trimmed, "#"):
			docLines = append(docLines, strings.TrimSpace(strings.TrimPrefix(trimmed, "#")))
		case strings.Contains(trimmed, "="):
			parts := strings.SplitN(trimmed, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			s.Options[key] = Option{
				Key:     key,
				Type:    TypeString,
				Default: val,
				Doc:     strings.Join(docLines, " "),
			}
			docLines = nil
		}
	}
	return s, sc.Err()
}

// MergeFrom returns a copy of s with docs and defaults taken from other when
// present, while preserving s's curated Type, Enum, and Repeatable fields. Keys
// only in other are added with their discovered (string) type.
func (s *Schema) MergeFrom(other *Schema) *Schema {
	out := &Schema{Options: make(map[string]Option, len(s.Options))}
	for k, o := range s.Options {
		out.Options[k] = o
	}
	for k, disc := range other.Options {
		if cur, ok := out.Options[k]; ok {
			if disc.Doc != "" {
				cur.Doc = disc.Doc
			}
			if disc.Default != "" {
				cur.Default = disc.Default
			}
			out.Options[k] = cur
		} else {
			out.Options[k] = disc
		}
	}
	return out
}
