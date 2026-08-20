package wss

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"wss/fingerprint"
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
	sourceFingerprint, err := fingerprint.Hash(sourceDir, fingerprint.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if want := sourceFingerprint + ".tar.gz"; archiveName != want {
		t.Fatalf("ArchiveReportIfNoVulnerabilities() filename = %q, want %q", archiveName, want)
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

func TestCheckReportCompliance(t *testing.T) {
	tests := []struct {
		name      string
		alert     string
		compliant bool
	}{
		{
			name:      "compliant",
			alert:     `{"libraries":[{"vulnerabilities":[]}]}`,
			compliant: true,
		},
		{
			name:      "not compliant",
			alert:     `{"libraries":[{"vulnerabilities":[{"name":"CVE-1"}]}]}`,
			compliant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportRoot := filepath.Join(t.TempDir(), "report")
			projectDir := filepath.Join(reportRoot, "project")
			if err := os.MkdirAll(projectDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(projectDir, "alert.json"), []byte(tt.alert), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := CheckReportCompliance(reportRoot, "project")
			if err != nil {
				t.Fatalf("CheckReportCompliance() error = %v", err)
			}
			if got != tt.compliant {
				t.Fatalf("CheckReportCompliance() = %v, want %v", got, tt.compliant)
			}
		})
	}
}

func TestArchiveSourceDoesNotRequireComplianceReport(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "package.txt"), []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}

	archiveName, err := ArchiveSource(sourceDir, tempDir)
	if err != nil {
		t.Fatalf("ArchiveSource() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, archiveName)); err != nil {
		t.Fatalf("archive does not exist: %v", err)
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

func TestReportArchiveNameUsesSourceFingerprint(t *testing.T) {
	const sourceFingerprint = "4d5e6f"
	const want = sourceFingerprint + ".tar.gz"
	if got := reportArchiveName(sourceFingerprint); got != want {
		t.Fatalf("reportArchiveName() = %q, want %q", got, want)
	}
}

func TestArchiveNameChangesWithSourceContent(t *testing.T) {
	tempDir := t.TempDir()
	reportRoot := filepath.Join(tempDir, "report")
	projectDir := filepath.Join(reportRoot, "safe-project")
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "alert.json"), []byte(`{"libraries":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	sourceFile := filepath.Join(sourceDir, "package.txt")
	if err := os.WriteFile(sourceFile, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}

	first, err := ArchiveReportIfNoVulnerabilities(reportRoot, "safe-project", sourceDir, tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourceFile, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	second, err := ArchiveReportIfNoVulnerabilities(reportRoot, "safe-project", sourceDir, tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("archive filename did not change after source content changed: %q", first)
	}
}
