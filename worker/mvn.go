package worker

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type Mvn struct{}

func (mvn Mvn) Download(destination string, packageName string, indexUrl string) string {
	var cmd string
	cmd = fmt.Sprintf("pip download %s --dest %s/%s/ %s", indexUrl, destination, packageName, packageName)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		panic(err)
	}
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
	// mvn dependency:copy-dependencies -DoutputDirectory=your-specific-directory
	dlprt := fmt.Sprintf("-DoutputDirectory=%s", downloadDestination)
	cmdArgs := []string{"dependency:copy-dependencies", dlprt}
	fmt.Println(cmdArgs)
	cmd := exec.Command("mvn", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mvn download failed: %w, output: %s", err, string(out))
	}
	return nil
}

func (mvn Mvn) Sync(targetUrl string, packageFile string) string {

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

func (mvn Mvn) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
