package worker

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UrlGet 結構體用於處理從 URL 獲取檔案的操作。
type UrlGet struct {
	Command string // 用於執行下載命令的指令 (例如 "wget" 或 "curl")。
}

// DownloadTask 結構體表示一個下載任務的詳細資訊。
type DownloadTask struct {
	Command             string // 下載命令。
	URL                 string // 要下載的檔案的 URL。
	Filename            string // 儲存下載檔案的名稱。
	DownloadDestination string // 下載檔案的目的地目錄。
}

// readURLsFromFile 從指定檔案路徑讀取 URL 列表。
// 它會逐行掃描檔案，忽略空行，並將以 "http" 或 "https" 開頭的行視為有效 URL。
//
// 參數:
//   - filePath: 包含 URL 列表的檔案路徑。
//
// 返回:
//   - []string: 從檔案中讀取到的 URL 字串切片。
//   - error: 如果無法開啟檔案或讀取時發生錯誤，返回錯誤；否則返回 nil。
func readURLsFromFile(filePath string) ([]string, error) {
	var urls []string
	var err error = nil

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("無法開啟文件: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if len(line) >= 4 && (line[:4] == "http" || line[:5] == "https") {
			urls = append(urls, line)
		} else {
			log.Printf("警告: 忽略無效的URL: %s\n", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return urls, fmt.Errorf("讀取文件時發生錯誤: %w", err)
	}

	return urls, nil
}

// setFilename 從 URL 中解析出檔案名稱並設定給 DownloadTask。
// 如果 URL 以斜線結尾或沒有明確的檔案名稱，則 Filename 將為空。
func (dt *DownloadTask) setFilename() {
	parts := strings.Split(dt.URL, "/")
	filename := parts[len(parts)-1]

	if filename == "" || filename == "/" {
		dt.Filename = ""
	}

	if strings.Contains(filename, ".") {
		dt.Filename = filename
	}
}

// Download 是一個佔位符函式，用於從 URL 下載指定套件。
// 目前未實作具體功能。
func (ug UrlGet) Download(destination string, packageName string, indexUrl string) string {
	var cmd string
	return string(cmd)
}

// SyncPackages 根據 requirements 檔案中的 URL 列表並行下載套件。
// 它會創建必要的報告目錄，讀取 URL，構建下載任務，然後使用 ParallelDownload 進行並行下載。
//
// 參數:
//   - destination: 要同步套件的目的地目錄名稱。
//   - requirementsFile: 包含要下載 URL 列表的檔案路徑。
//
// 返回:
//   - error: 如果發生任何錯誤，將返回一個錯誤。否則，返回 nil 表示成功。
func (ug UrlGet) SyncPackages(destination string, requirementsFile string) error {

	var err error = nil
	var downloadTasks []DownloadTask
	packageTmp := os.Getenv("package_tmp")
	reportTmp := os.Getenv("report_tmp")
	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
	reportDestination := fmt.Sprintf("%s/%s", reportTmp, destination)

	if err = os.MkdirAll(reportDestination, 0755); err != nil {
		return fmt.Errorf("Create Dir failed: %w", err)
	}
	urls, _ := readURLsFromFile(requirementsFile)

	for _, url := range urls {
		downloadTask := DownloadTask{
			Command:             os.Getenv("wget"),
			URL:                 url,
			DownloadDestination: downloadDestination,
		}
		downloadTask.setFilename()
		downloadTasks = append(downloadTasks, downloadTask)
	}

	concurrencyStr := os.Getenv("concurrency")
	concurrencyInt, _ := strconv.Atoi(concurrencyStr)
	ParallelDownload(downloadTasks, concurrencyInt)

	return err
}

// Sync 是一個佔位符函式，用於將檔案同步到目標 URL。
// 目前未實作具體功能。
func (ug UrlGet) Sync(targetUrl string, packageFile string) string {
	var bodyString string = ""
	return string(bodyString)

}

// Remove 是一個佔位符函式，用於刪除指定套件名稱對應的臨時目錄。
// 目前未實作具體功能。
func (ug UrlGet) Remove(packageName string) error {
	var err error
	return err

}

// DownloadFile 執行單個下載任務。
// 它會創建下載目錄，然後使用指定的命令和 URL 下載檔案到目的地。
//
// 參數:
//   - task: 要執行的下載任務。
//   - wg: 等待組，用於通知下載完成。
//   - errChan: 錯誤通道，用於傳遞下載過程中發生的錯誤。
func DownloadFile(task DownloadTask, wg *sync.WaitGroup, errChan chan error) {

	var err error = nil
	defer wg.Done()

	if err = os.MkdirAll(task.DownloadDestination, 0755); err != nil {
		errChan <- fmt.Errorf("Create directory failed: %w", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	dest := fmt.Sprintf("%s/%s", task.DownloadDestination, task.Filename)
	cmdArgs := []string{task.URL, "-o", dest}
	cmd := exec.CommandContext(ctx, task.Command, cmdArgs...)
	fmt.Println(cmd)
	// cmd := exec.CommandContext(ctx, os.Getenv("wget"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errChan <- fmt.Errorf("curl download failed: %w, output: %s", err, string(out))
	}
}

// ParallelDownload 並行執行多個下載任務，並限制最大並發數。
// 它使用 goroutine、WaitGroup 和帶緩衝的通道來管理並發下載和錯誤處理。
//
// 參數:
//   - tasks: 要執行的下載任務切片。
//   - maxConcurrency: 最大並發下載數。
//
// 返回:
//   - error: 如果任何下載任務失敗，返回第一個遇到的錯誤；否則返回 nil。
func ParallelDownload(tasks []DownloadTask, maxConcurrency int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	sem := make(chan struct{}, maxConcurrency) // Semaphore to limit concurrency

	for _, task := range tasks {
		wg.Add(1)

		go func(task DownloadTask) {
			sem <- struct{}{} // Acquire a slot
			DownloadFile(task, &wg, errChan)
			<-sem // Release the slot
		}(task)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	fmt.Println("test")
	return nil
}
