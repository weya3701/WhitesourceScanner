package worker

import (
	"fmt"
	"os"
	"os/exec"
)

type Mvn struct{}

func (mvn Mvn) Download(destination string, packageName string, indexUrl string) string {
	var out string = ""
	return string(out)
}

func (mvn Mvn) SyncPackages(destination string, requirementsFile string) error {
	packageTmp := os.Getenv("package_tmp")
	if packageTmp == "" {
		return fmt.Errorf("package_tmp is empty")

	}

	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
	if err := os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("Create Dir failed: %w", err)
	}
	outputDirectory := fmt.Sprintf("-DoutputDirectory=%s", downloadDestination)
	cmdArgs := []string{"dependency:copy-dependencies", "-f", requirementsFile, outputDirectory}
	fmt.Println(cmdArgs)
	cmd := exec.Command("mvn", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mvn download failed: %w, output: %s", err, string(out))
	}
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
