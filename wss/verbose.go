package wss

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

var verboseEnabled atomic.Bool

var verboseOutput = struct {
	sync.Mutex
	writer io.Writer
}{writer: os.Stdout}

// SetVerbose enables or disables detailed scan and report progress output.
func SetVerbose(enabled bool) {
	verboseEnabled.Store(enabled)
}

// IsVerbose reports whether detailed progress output is enabled.
func IsVerbose() bool {
	return verboseEnabled.Load()
}

// Verbosef writes a verbose message when verbose output is enabled.
func Verbosef(format string, args ...any) {
	if !IsVerbose() {
		return
	}
	verboseOutput.Lock()
	defer verboseOutput.Unlock()
	fmt.Fprintf(verboseOutput.writer, "[verbose] "+format+"\n", args...)
}

func getVerboseWriter() io.Writer {
	verboseOutput.Lock()
	defer verboseOutput.Unlock()
	return verboseOutput.writer
}
