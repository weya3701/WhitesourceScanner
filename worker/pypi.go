package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Pypi struct {
	Command string
}

func (py Pypi) Download(destination string, packageName string, indexUrl string) string {
	var cmd string
	cmd = fmt.Sprintf("pip download %s --dest %s/%s/ %s", indexUrl, destination, packageName, packageName)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println("Failed execute command")
	}
	return string(out)
}

// SyncPackages 函數用於從 PyPI (Python Package Index) 下載指定的 Python 包並同步到指定的目的地。
//
// 參數:
//   - destination: string -  下載包的目的地名稱，會被用來建立 download 和 report 目錄下的子目錄。
//   - requirementsFile: string - requirements.txt 文件的路徑，其中列出了要下載的 Python 包。
//
// 返回值:
//   - error: error - 如果發生錯誤，則返回錯誤信息；否則返回 nil。
func (py Pypi) SyncPackages(destination string, requirementsFile string) error {

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmdArgs := []string{"download", "-r", requirementsFile, "-d", downloadDestination}
	cmd := exec.CommandContext(ctx, py.Command, cmdArgs...)
	// cmd := exec.CommandContext(ctx, os.Getenv("pip"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip download failed: %w, output: %s", err, string(out))
	}

	return nil
}

func (py Pypi) Sync(targetUrl string, packageFile string) string {

	var body bytes.Buffer

	apiUrl := targetUrl
	file, err := os.Open(packageFile)
	if err != nil {
		fmt.Println("Failed to open file")
	}
	defer file.Close()

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, err := writer.CreateFormFile("file", filepath.Base(packageFile))
	if err != nil {
		fmt.Println("Failed to write")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Failed to copy file")
	}
	writer.Close()

	request, err := http.NewRequest("POST", apiUrl, requestBody)
	if err != nil {
		fmt.Println("Failed to send request")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.SetBasicAuth("admin", "admin")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Failed to send request")
	}
	defer response.Body.Close()

	_, err = io.Copy(&body, response.Body)
	// body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to copy file")
	}
	bodyString := body.String()
	return string(bodyString)

}

func (py Pypi) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
