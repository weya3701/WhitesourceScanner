package worker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadURLsFromFileFiltersInvalidLines(t *testing.T) {
	file := filepath.Join(t.TempDir(), "urls.txt")
	content := "https://example.com/a.zip\nhttp\nftp://example.com/file\n http://example.com/b.zip \n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	urls, err := readURLsFromFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("got %d URLs (%v), want 2", len(urls), urls)
	}
}

func TestParallelDownloadRejectsInvalidConcurrency(t *testing.T) {
	if err := ParallelDownload(nil, 0); err == nil {
		t.Fatal("expected invalid concurrency error")
	}
}

func TestSafeDestinationRejectsTraversal(t *testing.T) {
	for _, destination := range []string{"", "..", "../escape", "a/b"} {
		if _, err := safeDestination(t.TempDir(), destination); err == nil {
			t.Errorf("safeDestination accepted %q", destination)
		}
	}
}
