package worker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGradleSyncPackagesDoesNotModifyBuildFile(t *testing.T) {
	dir := t.TempDir()
	packageDir := filepath.Join(dir, "packages")
	reportDir := filepath.Join(dir, "reports")
	buildFile := filepath.Join(dir, "build.gradle")
	original := []byte("plugins { id 'java' }\n")
	if err := os.WriteFile(buildFile, original, 0644); err != nil {
		t.Fatal(err)
	}

	argsFile := filepath.Join(dir, "args.txt")
	command := filepath.Join(dir, "fake-gradle")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" >> \"$ARGS_FILE\"\n"
	if err := os.WriteFile(command, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ARGS_FILE", argsFile)
	t.Setenv("package_tmp", packageDir)
	t.Setenv("report_tmp", reportDir)

	if err := (Gradle{Command: command}).SyncPackages("project", buildFile); err != nil {
		t.Fatalf("SyncPackages() error = %v", err)
	}
	after, err := os.ReadFile(buildFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatalf("build file was modified: %q", after)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(args), "--init-script") || !strings.Contains(string(args), "downloadDependencies") {
		t.Fatalf("Gradle args did not use init script: %s", args)
	}
}
