package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateArguments(t *testing.T) {
	tests := []struct {
		name                                                                       string
		mode, packageName, projectName, application, packageType, requirementsFile string
		wantErr                                                                    bool
	}{
		{name: "cmd", mode: "cmd", packageName: "pkg", projectName: "project"},
		{name: "image", mode: "image", projectName: "project", application: "app"},
		{name: "invalid mode", mode: "unknown", wantErr: true},
		{name: "path traversal", mode: "cmd", packageName: "pkg", projectName: "../project", wantErr: true},
		{name: "missing image application", mode: "image", projectName: "project", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArguments(tt.mode, tt.packageName, tt.projectName, tt.application, tt.packageType, tt.requirementsFile)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateArguments() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateArgumentsReqfile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "requirements.txt")
	if err := os.WriteFile(file, []byte("example==1.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := validateArguments("reqfile", "pkg", "project", "", "pip", file); err != nil {
		t.Fatalf("valid reqfile arguments rejected: %v", err)
	}
}

func TestBatchRunnerStopsAfterFailure(t *testing.T) {
	runs := 0
	runner := NewBatchRunner([]BatchTask{
		{Name: "fail", Func: func() (bool, error) { runs++; return false, nil }},
		{Name: "must not run", Func: func() (bool, error) { runs++; return true, nil }},
	})
	if success, err := runner.Run(); success || err == nil {
		t.Fatalf("Run() = (%v, %v), want failure", success, err)
	}
	if runs != 1 {
		t.Fatalf("ran %d tasks, want 1", runs)
	}
}
