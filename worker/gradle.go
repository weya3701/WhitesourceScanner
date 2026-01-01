package worker

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Gradle struct{}

type ReplaceRule struct {
	Old string
	New string
}

func readFileContent(filePath string, rules []ReplaceRule) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var content strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}
		for _, rule := range rules {
			line = strings.ReplaceAll(line, rule.Old, rule.New)
		}

		content.WriteString(line)
	}
	return content.String(), nil
}

func appendToFile(filePath string, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func (gradle Gradle) Download(destination string, packageName string, indexUrl string) string {
	var cmd string = "Need to implement"
	return string(cmd)
}

// FIXME. 需新增./templates/build_tasks.gradled檔案
func (gradle Gradle) SyncPackages(destination string, requirementsFile string) error {

	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	tplFile := "./templates/build_tasks.gradle"
	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
	replaceRules := []ReplaceRule{
		{Old: "<destPath>", New: downloadDestination},
	}

	tplContent, err := readFileContent(tplFile, replaceRules)
	if err != nil {
		fmt.Println("Error reading tplFile:", err)
		return err
	}
	err = appendToFile(requirementsFile, tplContent)
	if err != nil {
		fmt.Println("Error appending to requirementsFile:", err)
		return err
	}
	if err := os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("Create Dir failed:%w", err)
	}
	// gradle -p /Users/ccxn/Desktop/demoGradle/ downloadDependencies
	cmdArgs := []string{"-p", "./", "downloadDependencies"}
	cmd := exec.Command("gradle", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gradle download failed: %w, output %s", err, string(out))
	}

	return nil
}

func (gradle Gradle) Sync(targetUrl string, packageFile string) string {
	var output string = "Need to implement"
	return string(output)

}

func (gradle Gradle) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
