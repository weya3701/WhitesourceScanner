package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Npm struct{}

func (npm Npm) Download(destination string, packageName string, indexUrl string) string {
	var cmd string = ""
	return string(cmd)
}

func (npm Npm) SyncPackages(destination string, requirementsFile string) error {

	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	reportTmp := os.Getenv("report_tmp")
	if packageTmp == "" {
		return fmt.Errorf("package_tmp is empty")
	}

	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
	reportDestination := fmt.Sprintf("%s/%s", reportTmp, destination)
	if err = os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("Create Dir failed: %w", err)
	}

	if err = os.MkdirAll(reportDestination, 0755); err != nil {
		return fmt.Errorf("Create Dir failed: %w", err)
	}
	copyCmd := []string{requirementsFile, downloadDestination}
	cpcmd := exec.Command("cp", copyCmd...)
	cpout, err := cpcmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Copy package.json failed: %w, output: %s", err, string(cpout))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmdArgs := []string{"install", "-prefix", downloadDestination}
	cmd := exec.CommandContext(ctx, os.Getenv("npm"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip download failed: %w, output: %s", err, string(out))
	}

	return nil
}

func (npm Npm) Sync(targetUrl string, packageFile string) string {
	var body string = ""
	return string(body)

}

func (npm Npm) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
