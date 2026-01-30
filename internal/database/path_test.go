package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	// Save and restore environment
	oldEnv := os.Getenv("FAZT_DB_PATH")
	defer os.Setenv("FAZT_DB_PATH", oldEnv)

	tests := []struct {
		name     string
		explicit string
		envValue string
		want     string
	}{
		{
			name:     "explicit path has highest priority",
			explicit: "/custom/path/data.db",
			envValue: "/env/path/data.db",
			want:     "/custom/path/data.db",
		},
		{
			name:     "env fallback when no explicit",
			explicit: "",
			envValue: "/env/path/data.db",
			want:     "/env/path/data.db",
		},
		{
			name:     "default when nothing set",
			explicit: "",
			envValue: "",
			want:     DefaultDBPath,
		},
		{
			name:     "explicit overrides env",
			explicit: "/explicit/data.db",
			envValue: "/env/data.db",
			want:     "/explicit/data.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("FAZT_DB_PATH", tt.envValue)
			got := ResolvePath(tt.explicit)
			if got != tt.want {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.explicit, got, tt.want)
			}
		})
	}
}

func TestResolvePath_ExpandsTilde(t *testing.T) {
	// Save and restore environment
	oldEnv := os.Getenv("FAZT_DB_PATH")
	defer os.Setenv("FAZT_DB_PATH", oldEnv)
	os.Setenv("FAZT_DB_PATH", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not get home dir")
	}

	tests := []struct {
		name     string
		explicit string
		want     string
	}{
		{
			name:     "expands tilde in explicit path",
			explicit: "~/fazt/data.db",
			want:     filepath.Join(home, "fazt/data.db"),
		},
		{
			name:     "absolute path unchanged",
			explicit: "/absolute/path/data.db",
			want:     "/absolute/path/data.db",
		},
		{
			name:     "relative path unchanged",
			explicit: "relative/data.db",
			want:     "relative/data.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePath(tt.explicit)
			if got != tt.want {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.explicit, got, tt.want)
			}
		})
	}
}

func TestResolvePath_EnvExpandsTilde(t *testing.T) {
	// Save and restore environment
	oldEnv := os.Getenv("FAZT_DB_PATH")
	defer os.Setenv("FAZT_DB_PATH", oldEnv)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not get home dir")
	}

	os.Setenv("FAZT_DB_PATH", "~/fazt/data.db")
	got := ResolvePath("")
	want := filepath.Join(home, "fazt/data.db")

	if got != want {
		t.Errorf("ResolvePath(\"\") with env=~/fazt/data.db = %q, want %q", got, want)
	}
}
