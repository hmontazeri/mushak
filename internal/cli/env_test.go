package cli

import (
	"fmt"
	"testing"
)

func TestUpdateEnvFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		updates  map[string]string
		expected string
	}{
		{
			name:     "empty content",
			content:  "",
			updates:  map[string]string{"FOO": "bar"},
			expected: "FOO=bar\n",
		},
		{
			name:     "replace existing",
			content:  "FOO=baz\n",
			updates:  map[string]string{"FOO": "bar"},
			expected: "FOO=bar\n",
		},
		{
			name:     "append new",
			content:  "FOO=baz\n",
			updates:  map[string]string{"BAR": "qux"},
			expected: "FOO=baz\nBAR=qux\n",
		},
		{
			name:     "mixed replace and append",
			content:  "A=1\nB=2\n",
			updates:  map[string]string{"A": "3", "C": "4"},
			expected: "A=3\nB=2\nC=4\n",
		},
		{
			name:    "preserve comments",
			content: "# this is a comment\nA=1\n",
			updates: map[string]string{"A": "2"},
			expected: "# this is a comment\nA=2\n",
		},
		{
			name:    "preserve empty lines",
			content: "A=1\n\nB=2\n",
			updates: map[string]string{"A": "3"},
			expected: "A=3\n\nB=2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateEnvFile(tt.content, tt.updates)
			if got != tt.expected {
				t.Errorf("updateEnvFile() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPluralizeEnv(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{count: 0, want: "s"},
		{count: 1, want: ""},
		{count: 2, want: "s"},
		{count: 10, want: "s"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.count), func(t *testing.T) {
			got := pluralizeEnv(tt.count)
			if got != tt.want {
				t.Errorf("pluralizeEnv(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}
