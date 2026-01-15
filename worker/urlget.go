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

type UrlGet struct{}

type DownloadTask struct {
	URL                 string
	Filename            string
	DownloadDestination string
}

// readURLsFromFile 讀取指定文件中的URL，並返回一個字符串切片
func readURLsFromFile(filePath string) ([]string, error) {
	var urls []string
	var err error = nil

	// 開啟文件
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

// FIXME. Need to implement.
// pip download from requirement file.
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

func DownloadFile(task DownloadTask, wg *sync.WaitGroup, errChan chan error) {

	var err error = nil
	defer wg.Done()

	if err = os.MkdirAll(task.DownloadDestination, 0755); err != nil {
		errChan <- fmt.Errorf("Create directory failed: %w", err)
		return
	}

	// Run curl download file command.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	dest := fmt.Sprintf("%s/%s", task.DownloadDestination, task.Filename)
	cmdArgs := []string{task.URL, "-o", dest}
	cmd := exec.CommandContext(ctx, os.Getenv("wget"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errChan <- fmt.Errorf("curl download failed: %w, output: %s", err, string(out))
	}
}

func ParallelDownload(tasks []DownloadTask, maxConcurrency int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	sem := make(chan int, maxConcurrency)
	// sem := make(chan struct{}, maxConcurrency)

	for _, task := range tasks {
		wg.Add(1)

		go func(task DownloadTask) {
			sem <- 0
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
