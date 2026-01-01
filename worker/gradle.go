package worker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	var cmd string
	cmd = fmt.Sprintf("pip download %s --dest %s/%s/ %s", indexUrl, destination, packageName, packageName)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
	return string(out)
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

	apiUrl := targetUrl
	file, err := os.Open(packageFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, err := writer.CreateFormFile("file", filepath.Base(packageFile))
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}
	writer.Close()

	request, err := http.NewRequest("POST", apiUrl, requestBody)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetBasicAuth("admin", "admin")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return string(body)

}

func (gradle Gradle) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
