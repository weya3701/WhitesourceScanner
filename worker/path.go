package worker

import (
	"fmt"
	"path/filepath"
	"strings"
)

func safeDestination(base, destination string) (string, error) {
	if base == "" {
		return "", fmt.Errorf("base directory is empty")
	}
	if destination == "" || destination == "." || destination == ".." || filepath.IsAbs(destination) || strings.ContainsAny(destination, `/\\`) {
		return "", fmt.Errorf("invalid destination %q", destination)
	}
	return filepath.Join(base, destination), nil
}
