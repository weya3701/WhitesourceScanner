package wss

import "testing"

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
