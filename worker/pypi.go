package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"wss/repositoryclient"
)

// Pypi 結構體用於處理與 Python PyPI 套件相關的操作。
type Pypi struct {
	Command string // 用於執行 pip 命令的指令。
}

// Download 從 PyPI 下載指定套件到目的地。
//
// 參數:
//   - destination: 下載套件的目標目錄。
//   - packageName: 要下載的套件名稱。
//   - indexUrl: PyPI 索引 URL (可選)。
//
// 返回:
//   - string: 命令的輸出結果。
func (py Pypi) Download(destination string, packageName string, indexUrl string) string {
	var cmd string
	cmd = fmt.Sprintf("pip download %s --dest %s/%s/ %s", indexUrl, destination, packageName, packageName)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println("Failed execute command")
	}
	return string(out)
}

// SyncPackages 根據 requirements 文件同步 PyPI 套件。
// 它會創建必要的下載和報告目錄，然後使用 pip download 命令下載套件。
//
// 參數:
//   - destination: 要同步套件的目的地目錄名稱 (例如 "package1")。此名稱將用作子目錄，在 package_tmp 和 report_tmp 目錄中創建。
//   - requirementsFile: 包含要下載套件的 requirements 檔案的路徑。
//
// 返回:
//   - error: 如果發生任何錯誤，將返回一個錯誤。否則，返回 nil 表示成功。
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

// Sync 將指定的套件檔案上傳到目標 URL。
// 它使用 multipart/form-data 格式發送 POST 請求。
//
// 參數:
//   - targetUrl: 上傳目標的 URL。
//   - packageFile: 要上傳的套件檔案的路徑。
//
// 返回:
//   - string: 伺服器回應的主體內容。
func (py Pypi) Sync(targetUrl string, packageFile string) string {

	var body bytes.Buffer
	var err error = nil
	var bodyString string = ""

	apiUrl := targetUrl
	file, err := os.Open(packageFile)
	if err != nil {
		log.Println("Failed to open file")
	}
	defer file.Close()

	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, err := writer.CreateFormFile("file", filepath.Base(packageFile))
	if err != nil {
		log.Println("Failed to write")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Println("Failed to copy file")
	}
	writer.Close()

	request, err := http.NewRequest("POST", apiUrl, requestBody)
	if err != nil {
		log.Println("Failed to send request")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		log.Println("Failed to send request")
	}
	defer response.Body.Close()

	_, err = io.Copy(&body, response.Body)
	// body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Failed to copy file")
	}
	bodyString = string(body.String())
	return bodyString

}

// Publish 將位於 packageDirPath 的套件發佈到 PyPI 儲存庫。
// 它使用 `twine upload` 命令，並透過環境變數處理認證。
// packageDirPath 應該包含要上傳的 `.whl` 或 `.tar.gz` 檔案。
func (py Pypi) Publish(conn *repositoryclient.RepositoryConnection, packageDirPath string) error {
	// 假設 twine 命令已安裝並在 PATH 中
	// 假設 TWINE_USERNAME 和 TWINE_PASSWORD 已在環境變數中設置，或者通過 .pypirc 配置
	// Alternatively, pass --username and --password as arguments, but env vars are safer.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 設置 Twine 認證環境變數 (如果連接資訊提供)
	// 如果 conn.Username 和 conn.PAT 是空字串，這將不會覆蓋已存在的環境變數
	if conn.Username != "" {
		os.Setenv("TWINE_USERNAME", conn.Username)
	}
	if conn.PAT != "" {
		os.Setenv("TWINE_PASSWORD", conn.PAT)
	}

	// 尋找要上傳的檔案模式，通常是 *.whl 或 *.tar.gz
	// `packageDirPath` 應該是包含這些檔案的目錄
	uploadPattern := filepath.Join(packageDirPath, "*")

	// `twine upload --repository-url <URL> <files...>`
	cmdArgs := []string{"upload", "--repository-url", conn.Url, uploadPattern}

	cmd := exec.CommandContext(ctx, "twine", cmdArgs...) // 假設 twine 在 PATH 中
	cmd.Dir = packageDirPath                             // 在 packageDirPath 執行，以便 `*` 能正確展開

	// 捕獲標準輸出和錯誤輸出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Printf("Executing twine: %s %v\n", "twine", cmdArgs)
	log.Printf("Twine command output:\n%s\n%s", stdout.String(), stderr.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("twine upload failed: %w\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	log.Printf("Successfully published packages from %s to %s", packageDirPath, conn.Url)
	return nil
}

// Remove 刪除指定套件名稱對應的臨時目錄。
//
// 參數:
//   - packageName: 要刪除的套件名稱。
//
// 返回:
//   - error: 如果刪除失敗，返回錯誤；否則返回 nil。
func (py Pypi) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
