package config

// Get returns the value of the first KeyValue line with the given key.
func (d *Document) Get(key string) (string, bool) {
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			return l.Value, true
		}
	}
	return "", false
}

// GetAll returns the values of every KeyValue line with the given key, in order.
func (d *Document) GetAll(key string) []string {
	var out []string
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			out = append(out, l.Value)
		}
	}
	return out
}

// Set updates the first existing line with key, or appends a new line.
func (d *Document) Set(key, value string) {
	for i := range d.Lines {
		if d.Lines[i].Kind == KindKeyValue && d.Lines[i].Key == key {
			d.Lines[i].Value = value
			d.Lines[i].dirty = true
			return
		}
	}
	d.append(key, value)
}

// RemoveAll deletes every KeyValue line with the given key.
func (d *Document) RemoveAll(key string) {
	kept := d.Lines[:0]
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			continue
		}
		kept = append(kept, l)
	}
	d.Lines = kept
}

// SetRepeatable replaces all lines for key with one line per value, appended at
// the position of the former first occurrence (or end of file if none existed).
func (d *Document) SetRepeatable(key string, values []string) {
	insertAt := -1
	kept := make([]Line, 0, len(d.Lines))
	for _, l := range d.Lines {
		if l.Kind == KindKeyValue && l.Key == key {
			if insertAt == -1 {
				insertAt = len(kept)
			}
			continue
		}
		kept = append(kept, l)
	}
	newLines := make([]Line, 0, len(values))
	for _, v := range values {
		newLines = append(newLines, Line{Kind: KindKeyValue, Key: key, Value: v, dirty: true})
	}
	if insertAt == -1 {
		insertAt = len(kept)
	}
	if d.trailingNewline || len(d.Lines) == 0 {
		d.trailingNewline = true
	}
	d.Lines = append(kept[:insertAt:insertAt], append(newLines, kept[insertAt:]...)...)
}

func (d *Document) append(key, value string) {
	d.Lines = append(d.Lines, Line{Kind: KindKeyValue, Key: key, Value: value, dirty: true})
	d.trailingNewline = true
}
