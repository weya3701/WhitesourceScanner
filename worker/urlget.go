package worker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UrlGet struct {
	Command string
}

type DownloadTask struct {
	Command             string
	URL                 string
	Filename            string
	DownloadDestination string
}

// readURLsFromFile 從給定的文件路徑讀取URL列表。
//
// 參數:
//   - filePath: 包含URL的文件路徑。每行應包含一個URL。
//
// 返回值:
//   - []string: 包含從文件中讀取的有效URL的切片。
//   - error: 如果發生錯誤，則返回錯誤; 否則為nil。
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
			fmt.Printf("警告: 忽略無效的URL: %s\n", line) // 打印警告，方便調試
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("讀取文件時發生錯誤: %w", err)
	}

	return urls, nil
}

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

func (ug UrlGet) Download(destination string, packageName string, indexUrl string) string {
	var cmd string
	return string(cmd)
}

// SyncPackages 函數用於同步下載由 requirementsFile 指定的 URL 列表中指定的软件包，到指定的目標目錄。
// 它首先創建用於存儲報告和下載文件的臨時目錄，然後從 requirementsFile 中讀取 URL 列表。
// 接著，它為每個 URL 創建一個 DownloadTask，設定下載文件的文件名和目標路徑，並將其添加到 downloadTasks 切片中。
// 最後，它使用 ParallelDownload 函數并行下載所有任務，並根據 concurrency 环境变量控制并行度。
//
// 參數:
//   - destination: 目標目錄的名稱，用於構建存儲下載包的目錄和報告目錄。
//   - requirementsFile: 包含要下載的 URL 列表的文件路徑。
//
// 返回值:
//   - error: 如果發生錯誤，則返回錯誤信息；否則返回 nil。
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

func (ug UrlGet) Sync(targetUrl string, packageFile string) string {
	var bodyString string = ""
	return string(bodyString)

}

func (ug UrlGet) Remove(packageName string) error {
	var err error
	return err

}

// DownloadFile 函數使用 wget 從給定的 URL 下載檔案到指定的目的地。
//
// 參數:
//   - task：DownloadTask 結構體，包含下載任務的詳細資訊，如 URL、檔案名稱和下載目的地。
//   - wg：sync.WaitGroup 指標，用於等待 goroutine 完成。
//   - errChan：錯誤通道，用於報告下載過程中發生的錯誤。
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

// ParallelDownload 并行下载多个文件。
//
// 参数:
//   - tasks: 包含需要下载任务的切片，每个任务定义了下载文件的信息（例如URL，保存路径等）。
//   - maxConcurrency: 并发下载的最大数量。限制同时运行的 goroutine 的数量，防止过度消耗资源。
//
// 返回值:
//   - error:  如果下载过程中发生任何错误，则返回该错误。如果所有下载任务都成功完成，则返回 nil。
//
// 流程:
// 1. 使用 sync.WaitGroup 来等待所有下载任务完成。
// 2. 创建一个带缓冲的 channel errChan 来收集下载过程中可能出现的错误。缓冲大小设置为任务总数，确保不会因为 channel 满而阻塞。
// 3. 创建一个带缓冲的 semaphore channel sem，用于限制并发数量。其缓冲大小设置为 maxConcurrency。
// 4. 遍历 tasks 切片，对每个任务启动一个 goroutine 进行下载:
//   - 在每个 goroutine 中，首先通过 `sem <- struct{}{}` 获取 semaphore，表示占用一个并发槽位。
//   - 调用 DownloadFile 函数执行实际的下载操作，并将 sync.WaitGroup、errChan 传递进去，以方便同步和错误处理。
//   - 下载完成后，使用 `<-sem` 释放 semaphore，释放并发槽位。
//
// 5. 启动一个 goroutine，在所有下载任务完成后关闭 errChan，通知错误收集 goroutine 停止接收错误。
// 6. 循环遍历 errChan，检查是否有错误发生:
//   - 如果从 errChan 接收到错误，则立即返回该错误。
//   - 如果 errChan 被关闭且没有错误，则表示所有下载都成功完成，返回 nil。
func ParallelDownload(tasks []DownloadTask, maxConcurrency int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	sem := make(chan struct{}, maxConcurrency)

	for _, task := range tasks {
		wg.Add(1)

		go func(task DownloadTask) {
			sem <- struct{}{}
			DownloadFile(task, &wg, errChan)
			<-sem
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

	return nil
}
