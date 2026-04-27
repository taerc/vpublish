package model

import "testing"

func TestIsValidFeatureType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"debug type", FeatureTypeDebug, true},
		{"release type", FeatureTypeRelease, true},
		{"demo type", FeatureTypeDemo, true},
		{"string literal debug", "debug", true},
		{"string literal release", "release", true},
		{"string literal demo", "demo", true},
		{"invalid type", "invalid", false},
		{"empty string", "", false},
		{"uppercase DEBUG", "DEBUG", false},
		{"uppercase RELEASE", "RELEASE", false},
		{"whitespace", " debug ", false},
		{"partial match", "rel", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidFeatureType(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidFeatureType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFeatureTypeConstants(t *testing.T) {
	if FeatureTypeDebug != "debug" {
		t.Errorf("FeatureTypeDebug = %q, want %q", FeatureTypeDebug, "debug")
	}
	if FeatureTypeRelease != "release" {
		t.Errorf("FeatureTypeRelease = %q, want %q", FeatureTypeRelease, "release")
	}
	if FeatureTypeDemo != "demo" {
		t.Errorf("FeatureTypeDemo = %q, want %q", FeatureTypeDemo, "demo")
	}
}
