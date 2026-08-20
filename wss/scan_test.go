package wss

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParserEnvUsesDefaultsWhenFileIsMissing(t *testing.T) {
	t.Setenv("MEND_API_KEY", "api-key")
	t.Setenv("MEND_USER_KEY", "user-key")
	t.Setenv("MEND_PRODUCT_NAME", "product-name")
	t.Setenv("MEND_PRODUCT_TOKEN", "product-token")
	var config WhiteSourceEnv
	missing := filepath.Join(t.TempDir(), "conf.yaml")

	if err := config.ParserEnv(missing); err != nil {
		t.Fatalf("ParserEnv() error = %v", err)
	}
	want := defaultWhiteSourceEnv
	want.ApiKey = "api-key"
	want.UserKey = "user-key"
	want.ProductName = "product-name"
	want.ProductToken = "product-token"
	if config != want {
		t.Fatalf("ParserEnv() = %#v, want %#v", config, want)
	}
}

func TestParserEnvRequiresMendEnvironmentWhenFileIsMissing(t *testing.T) {
	t.Setenv("MEND_API_KEY", "")
	t.Setenv("MEND_USER_KEY", "")
	t.Setenv("MEND_PRODUCT_NAME", "")
	t.Setenv("MEND_PRODUCT_TOKEN", "")
	var config WhiteSourceEnv
	missing := filepath.Join(t.TempDir(), "conf.yaml")

	if err := config.ParserEnv(missing); err == nil {
		t.Fatal("ParserEnv() accepted missing Mend environment variables")
	}
}

func TestParserEnvRequiresMendProductNameWhenFileIsMissing(t *testing.T) {
	t.Setenv("MEND_API_KEY", "api-key")
	t.Setenv("MEND_USER_KEY", "user-key")
	t.Setenv("MEND_PRODUCT_NAME", "")
	t.Setenv("MEND_PRODUCT_TOKEN", "product-token")
	var config WhiteSourceEnv
	missing := filepath.Join(t.TempDir(), "conf.yaml")

	err := config.ParserEnv(missing)
	if err == nil || !strings.Contains(err.Error(), "MEND_PRODUCT_NAME") {
		t.Fatalf("ParserEnv() error = %v, want missing MEND_PRODUCT_NAME", err)
	}
}

func TestUnifiedAgentCommandArgsWithConfigFile(t *testing.T) {
	got := unifiedAgentCommandArgs("/opt/wss-agent.jar", "/src/project", "./config/custom.config")
	want := []string{
		"java", "-jar", "/opt/wss-agent.jar",
		"-d", "/src/project",
		"-c", "./config/custom.config",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unifiedAgentCommandArgs() = %q, want %q", got, want)
	}
}

func TestUnifiedAgentCommandArgsWithoutConfigFile(t *testing.T) {
	got := unifiedAgentCommandArgs("/opt/wss-agent.jar", "/src/project", "")
	want := []string{"java", "-jar", "/opt/wss-agent.jar", "-d", "/src/project"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unifiedAgentCommandArgs() = %q, want %q", got, want)
	}
}
