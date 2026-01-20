package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Mvn struct{}

func (mvn Mvn) Download(destination string, packageName string, indexUrl string) string {
	var out string = ""
	return string(out)
}

func (mvn Mvn) SyncPackages(destination string, requirementsFile string) error {
	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	reportTmp := os.Getenv("report_tmp")
	if packageTmp == "" {
		return fmt.Errorf("environment variable 'package_tmp' is not set")
	}

	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		return fmt.Errorf("requirements file not found: %s", requirementsFile)
	}

	downloadDestination := filepath.Join(packageTmp, destination)
	reportDestination := filepath.Join(reportTmp, destination)

	if err = os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("failed to create download dir: %w", err)
	}

	if err = os.MkdirAll(reportDestination, 0755); err != nil {
		return fmt.Errorf("failed to create download dir: %w", err)
	} else {
		fmt.Printf("create report directory %s successful.", reportDestination)
	}

	cmdArgs := []string{
		"-B",
		"dependency:copy-dependencies",
		"-f", requirementsFile,
		fmt.Sprintf("-Dmaven.repo.local=%s", downloadDestination),
		"-DincludeScope=runtime",
		"-U",
	}

	fmt.Printf("Executing Maven: mvn %v\n", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // 設定 10 分鐘超時
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Getenv("mvn"), cmdArgs...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mvn execution failed: %w\nOutput:\n%s", err, string(out))
	}

	dependenciesTreeFile := fmt.Sprintf("%s/dependenciesTree.txt", reportDestination)
	cmds := []string{"dependency:tree", "-f", requirementsFile, fmt.Sprintf("-Dmaven.repo.local=%s", downloadDestination), "-DincludeScope=runtime", "-U"}
	GetDependenciesTree(dependenciesTreeFile, os.Getenv("mvn"), cmds)

	return nil
}

func (mvn Mvn) Sync(targetUrl string, packageFile string) string {

	var body string = ""
	return string(body)

}

func (mvn Mvn) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
