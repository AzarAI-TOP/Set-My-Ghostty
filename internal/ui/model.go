package ui

import (
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/config"
	"github.com/AzarAI-TOP/Set-My-Ghostty/internal/schema"
)

// Model bridges a parsed config Document and a Schema to the widgets. Edits are
// staged in pending maps and flushed into the Document by Apply.
type Model struct {
	Doc    *config.Document
	Schema *schema.Schema

	pendingScalar map[string]string
	pendingList   map[string][]string
}

// NewModel builds a model over a document and schema.
func NewModel(doc *config.Document, s *schema.Schema) *Model {
	return &Model{
		Doc:           doc,
		Schema:        s,
		pendingScalar: map[string]string{},
		pendingList:   map[string][]string{},
	}
}

// Value returns the current value for a scalar key: pending edit, else the
// document value, else the schema default, else "".
func (m *Model) Value(key string) string {
	if v, ok := m.pendingScalar[key]; ok {
		return v
	}
	if v, ok := m.Doc.Get(key); ok {
		return v
	}
	if opt, ok := m.Schema.Options[key]; ok {
		return opt.Default
	}
	return ""
}

// SetValue stages a scalar edit. If the new value equals the document value the
// pending edit is cleared (no-op edits do not make the model dirty).
func (m *Model) SetValue(key, value string) {
	if cur, ok := m.Doc.Get(key); ok && cur == value {
		delete(m.pendingScalar, key)
		return
	}
	m.pendingScalar[key] = value
}

// List returns the current values for a repeatable key (pending else document).
func (m *Model) List(key string) []string {
	if v, ok := m.pendingList[key]; ok {
		return v
	}
	return m.Doc.GetAll(key)
}

// SetList stages an edit to a repeatable key.
func (m *Model) SetList(key string, values []string) {
	m.pendingList[key] = values
}

// Dirty reports whether there are unsaved edits.
func (m *Model) Dirty() bool {
	return len(m.pendingScalar) > 0 || len(m.pendingList) > 0
}

// Apply flushes all pending edits into the document and clears them.
func (m *Model) Apply() {
	for k, v := range m.pendingScalar {
		m.Doc.Set(k, v)
	}
	for k, vs := range m.pendingList {
		m.Doc.SetRepeatable(k, vs)
	}
	m.pendingScalar = map[string]string{}
	m.pendingList = map[string][]string{}
}
