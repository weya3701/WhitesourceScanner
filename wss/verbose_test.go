package wss

import (
	"bytes"
	"strings"
	"testing"
)

func captureVerboseOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	previousEnabled := IsVerbose()
	verboseOutput.Lock()
	previousWriter := verboseOutput.writer
	verboseOutput.writer = &output
	verboseOutput.Unlock()
	t.Cleanup(func() {
		SetVerbose(previousEnabled)
		verboseOutput.Lock()
		verboseOutput.writer = previousWriter
		verboseOutput.Unlock()
	})
	return &output
}

func TestVerbosefRespectsSetting(t *testing.T) {
	output := captureVerboseOutput(t)

	SetVerbose(false)
	Verbosef("hidden")
	if output.Len() != 0 {
		t.Fatalf("disabled verbose output = %q, want empty", output.String())
	}

	SetVerbose(true)
	Verbosef("project=%s status=%s", "example", "SUCCESS")
	got := output.String()
	if !strings.Contains(got, "[verbose]") || !strings.Contains(got, "project=example status=SUCCESS") {
		t.Fatalf("verbose output = %q, want detailed status", got)
	}
}
