package wss

import (
	"path/filepath"
	"testing"
)

func TestParserEnvUsesDefaultsWhenFileIsMissing(t *testing.T) {
	t.Setenv("MEND_API_KEY", "api-key")
	t.Setenv("MEND_USER_KEY", "user-key")
	t.Setenv("MEND_PRODUCT_TOKEN", "product-token")
	var config WhiteSourceEnv
	missing := filepath.Join(t.TempDir(), "conf.yaml")

	if err := config.ParserEnv(missing); err != nil {
		t.Fatalf("ParserEnv() error = %v", err)
	}
	want := defaultWhiteSourceEnv
	want.ApiKey = "api-key"
	want.UserKey = "user-key"
	want.ProductToken = "product-token"
	if config != want {
		t.Fatalf("ParserEnv() = %#v, want %#v", config, want)
	}
}

func TestParserEnvRequiresMendEnvironmentWhenFileIsMissing(t *testing.T) {
	t.Setenv("MEND_API_KEY", "")
	t.Setenv("MEND_USER_KEY", "")
	t.Setenv("MEND_PRODUCT_TOKEN", "")
	var config WhiteSourceEnv
	missing := filepath.Join(t.TempDir(), "conf.yaml")

	if err := config.ParserEnv(missing); err == nil {
		t.Fatal("ParserEnv() accepted missing Mend environment variables")
	}
}
