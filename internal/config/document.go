// Package config parses, edits, and serializes ghostty config files while
// preserving comments, blank lines, and the ordering of unedited lines.
package config

import (
	"bytes"
	"strings"
)

// LineKind classifies a line in a ghostty config file.
type LineKind int

const (
	KindBlank LineKind = iota
	KindComment
	KindKeyValue
)

// Line is one physical line of the config file.
type Line struct {
	Kind  LineKind
	Key   string // KindKeyValue only
	Value string // KindKeyValue only
	Raw   string // original text; used verbatim unless dirty is set
	dirty bool   // when true, Bytes() regenerates from Key/Value
}

// Document is an ordered list of lines plus trailing-newline state.
type Document struct {
	Lines           []Line
	trailingNewline bool
}

// Parse reads config bytes into a Document.
func Parse(b []byte) *Document {
	d := &Document{}
	s := string(b)
	if s == "" {
		return d
	}
	d.trailingNewline = strings.HasSuffix(s, "\n")
	body := s
	if d.trailingNewline {
		body = strings.TrimSuffix(s, "\n")
	}
	for _, raw := range strings.Split(body, "\n") {
		d.Lines = append(d.Lines, classify(raw))
	}
	return d
}

func classify(raw string) Line {
	trimmed := strings.TrimSpace(raw)
	switch {
	case trimmed == "":
		return Line{Kind: KindBlank, Raw: raw}
	case strings.HasPrefix(trimmed, "#"):
		return Line{Kind: KindComment, Raw: raw}
	case strings.Contains(raw, "="):
		parts := strings.SplitN(raw, "=", 2)
		return Line{
			Kind:  KindKeyValue,
			Key:   strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
			Raw:   raw,
		}
	default:
		// A line with no '=' and not a comment: keep verbatim, treat as comment-like.
		return Line{Kind: KindComment, Raw: raw}
	}
}

func (l Line) serialize() string {
	if l.Kind == KindKeyValue && l.dirty {
		return l.Key + " = " + l.Value
	}
	return l.Raw
}

// Bytes serializes the document back to config bytes.
func (d *Document) Bytes() []byte {
	if len(d.Lines) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for i, l := range d.Lines {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(l.serialize())
	}
	if d.trailingNewline {
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}
