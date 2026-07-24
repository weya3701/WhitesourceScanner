package wss

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestArchiveReportIfNoVulnerabilities(t *testing.T) {
	tempDir := t.TempDir()
	reportRoot := filepath.Join(tempDir, "report")
	projectDir := filepath.Join(reportRoot, "safe-project")
	sourceDir := filepath.Join(tempDir, "packages", "source-package")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	alert := `{"libraries":[{"name":"safe","vulnerabilities":[]}]}`
	if err := os.WriteFile(filepath.Join(projectDir, "alert.json"), []byte(alert), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "package.txt"), []byte("source package"), 0644); err != nil {
		t.Fatal(err)
	}

	archiveName, err := ArchiveReportIfNoVulnerabilities(reportRoot, "safe-project", sourceDir, tempDir)
	if err != nil {
		t.Fatalf("ArchiveReportIfNoVulnerabilities() error = %v", err)
	}
	if archiveName == "" {
		t.Fatal("ArchiveReportIfNoVulnerabilities() filename is empty")
	}

	file, err := os.Open(filepath.Join(tempDir, archiveName))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	foundSource := false
	foundAlert := false
	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}
		if header.Name == "source-package/package.txt" {
			foundSource = true
		}
		if header.Name == "safe-project/alert.json" {
			foundAlert = true
		}
	}
	if !foundSource {
		t.Fatal("archive does not contain source-package/package.txt")
	}
	if foundAlert {
		t.Fatal("archive unexpectedly contains report alert.json")
	}
}

func TestArchiveReportWithVulnerabilitiesDoesNothing(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "report", "risky-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	alert := `{"libraries":[{"vulnerabilities":[{"name":"CVE-1","severity":"HIGH"}]}]}`
	if err := os.WriteFile(filepath.Join(projectDir, "alert.json"), []byte(alert), 0644); err != nil {
		t.Fatal(err)
	}

	archiveName, err := ArchiveReportIfNoVulnerabilities(
		filepath.Join(tempDir, "report"),
		"risky-project",
		filepath.Join(tempDir, "missing-source"),
		tempDir,
	)
	if err != nil {
		t.Fatalf("ArchiveReportIfNoVulnerabilities() error = %v", err)
	}
	if archiveName != "" {
		t.Fatalf("ArchiveReportIfNoVulnerabilities() filename = %q, want empty", archiveName)
	}
	archives, err := filepath.Glob(filepath.Join(tempDir, "*.tar.gz"))
	if err != nil {
		t.Fatal(err)
	}
	if len(archives) != 0 {
		t.Fatalf("unexpected archives = %v", archives)
	}
}

func TestArchiveReportRejectsMissingLibraries(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "report", "invalid-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "alert.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := ArchiveReportIfNoVulnerabilities(
		filepath.Join(tempDir, "report"),
		"invalid-project",
		filepath.Join(tempDir, "missing-source"),
		tempDir,
	); err == nil {
		t.Fatal("ArchiveReportIfNoVulnerabilities() accepted a report without libraries")
	}
}

func TestReportArchiveNameUsesProjectNameAndTimestampHash(t *testing.T) {
	createdAt := time.Date(2026, time.July, 23, 14, 5, 6, 123456789, time.FixedZone("CST", 8*60*60))
	input := "safe-project" + "20260723140506.123456789"
	want := fmt.Sprintf("%x.tar.gz", sha256.Sum256([]byte(input)))

	if got := reportArchiveName("safe-project", createdAt); got != want {
		t.Fatalf("reportArchiveName() = %q, want %q", got, want)
	}
}
