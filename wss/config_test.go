package wss

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRuntimeConfigRejectsInvalidConcurrency(t *testing.T) {
	t.Setenv("concurrency", "0")
	if _, err := LoadRuntimeConfig(); err == nil {
		t.Fatal("expected invalid concurrency error")
	}
}

func TestRuntimeConfigImageNeedsNoScannerEnvironment(t *testing.T) {
	if err := (RuntimeConfig{}).Validate("image", ""); err != nil {
		t.Fatalf("image config validation failed: %v", err)
	}
}

func TestLoadEnvironmentUsesDefaultsWhenFileIsMissing(t *testing.T) {
	const key = "settings_file"
	previous, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})

	if err := LoadEnvironment(filepath.Join(t.TempDir(), ".env")); err != nil {
		t.Fatalf("LoadEnvironment() error = %v", err)
	}
	if got := os.Getenv(key); got != defaultEnvironment[key] {
		t.Fatalf("%s = %q, want %q", key, got, defaultEnvironment[key])
	}
}

func TestLoadEnvironmentLoadsExistingFile(t *testing.T) {
	const key = "WHITESOURCE_SCANNER_ENV_TEST"
	previous, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, previous)
		} else {
			_ = os.Unsetenv(key)
		}
	})

	file := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(file, []byte(key+"=from-file\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := LoadEnvironment(file); err != nil {
		t.Fatalf("LoadEnvironment() error = %v", err)
	}
	if got := os.Getenv(key); got != "from-file" {
		t.Fatalf("%s = %q, want from-file", key, got)
	}
}
