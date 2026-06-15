package schema

import "testing"

func TestStaticHasCoreKeys(t *testing.T) {
	s := Static()
	for _, key := range []string{"theme", "font-family", "font-size", "background-opacity", "keybind"} {
		opt, ok := s.Options[key]
		if !ok {
			t.Errorf("static schema missing %q", key)
			continue
		}
		if opt.Key != key {
			t.Errorf("option key mismatch: %q != %q", opt.Key, key)
		}
	}
}

func TestStaticTypesAndRepeatable(t *testing.T) {
	s := Static()
	if s.Options["background-opacity"].Type != TypeFloat {
		t.Errorf("background-opacity should be float")
	}
	if !s.Options["keybind"].Repeatable {
		t.Errorf("keybind should be repeatable")
	}
	if s.Options["font-size"].Type != TypeInt {
		t.Errorf("font-size should be int")
	}
}
