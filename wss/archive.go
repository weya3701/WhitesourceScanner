package wss

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"wss/fingerprint"
)

type alertInventory struct {
	Libraries *[]alertLibrary `json:"libraries"`
}

type alertLibrary struct {
	Vulnerabilities []json.RawMessage `json:"vulnerabilities"`
}

// CheckReportCompliance reports whether alert.json contains no vulnerabilities.
func CheckReportCompliance(reportRoot, projectName string) (bool, error) {
	reportDir := filepath.Join(reportRoot, projectName)
	alertFile := filepath.Join(reportDir, "alert.json")

	input, err := os.Open(alertFile)
	if err != nil {
		return false, fmt.Errorf("open alert report %s: %w", alertFile, err)
	}
	defer input.Close()

	var inventory alertInventory
	if err := json.NewDecoder(input).Decode(&inventory); err != nil {
		return false, fmt.Errorf("decode alert report %s: %w", alertFile, err)
	}
	if inventory.Libraries == nil {
		return false, fmt.Errorf("alert report %s does not contain libraries", alertFile)
	}
	for _, library := range *inventory.Libraries {
		if len(library.Vulnerabilities) > 0 {
			return false, nil
		}
	}
	return true, nil
}

// ArchiveSource packages sourceDir into workDir. The archive filename is
// derived from the source directory fingerprint.
func ArchiveSource(sourceDir, workDir string) (string, error) {
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		return "", fmt.Errorf("inspect scan source %s: %w", sourceDir, err)
	}
	if !sourceInfo.IsDir() {
		return "", fmt.Errorf("scan source %s is not a directory", sourceDir)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("create archive directory %s: %w", workDir, err)
	}
	sourceFingerprint, err := fingerprint.Hash(sourceDir, fingerprint.Options{})
	if err != nil {
		return "", fmt.Errorf("fingerprint scan source %s: %w", sourceDir, err)
	}
	archiveName := reportArchiveName(sourceFingerprint)
	archivePath := filepath.Join(workDir, archiveName)
	tempFile, err := os.CreateTemp(workDir, ".source-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("create temporary report archive: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	archiveRoot := filepath.Base(filepath.Clean(sourceDir))
	if err := writeTarGz(tempFile, sourceDir, archiveRoot); err != nil {
		tempFile.Close()
		return "", err
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close report archive: %w", err)
	}
	if err := os.Rename(tempPath, archivePath); err != nil {
		return "", fmt.Errorf("save report archive %s: %w", archivePath, err)
	}
	return archiveName, nil
}

// ArchiveReportIfNoVulnerabilities packages the scanned source directory when
// alert.json contains no vulnerabilities. It keeps the original combined
// behavior for callers that still need it.
func ArchiveReportIfNoVulnerabilities(reportRoot, projectName, sourceDir, workDir string) (string, error) {
	compliant, err := CheckReportCompliance(reportRoot, projectName)
	if err != nil || !compliant {
		return "", err
	}
	return ArchiveSource(sourceDir, workDir)
}

func reportArchiveName(sourceFingerprint string) string {
	return sourceFingerprint + ".tar.gz"
}

func writeTarGz(output io.Writer, sourceDir, archiveRoot string) error {
	gzipWriter := gzip.NewWriter(output)
	tarWriter := tar.NewWriter(gzipWriter)

	walkErr := filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		archiveName := archiveRoot
		if relativePath != "." {
			archiveName = filepath.Join(archiveRoot, relativePath)
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(archiveName)
		if info.IsDir() {
			header.Name += "/"
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tarWriter, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if walkErr != nil {
		tarWriter.Close()
		gzipWriter.Close()
		return fmt.Errorf("archive scan source %s: %w", sourceDir, walkErr)
	}
	if err := tarWriter.Close(); err != nil {
		gzipWriter.Close()
		return fmt.Errorf("finish tar archive: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("finish gzip archive: %w", err)
	}
	return nil
}
