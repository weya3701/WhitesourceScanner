package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Worker 介面定義了處理套件的通用操作。
type Worker interface {
	// Download 從指定索引 URL 下載套件到目標目錄。
	Download(destination string, packageName string, indexUrl string) string
	// Sync 將套件檔案同步到目標 URL。
	Sync(targetUrl string, packageFile string) string
	// Remove 刪除指定路徑的套件。
	Remove(fullPath string) error
	// SyncPackages 根據需求檔案同步套件。
	SyncPackages(destination string, requirementsFile string) error
}

// WorkerHandler 結構體封裝了一個 Worker 介面實例。
type WorkerHandler struct {
	worker Worker // 實際執行套件操作的 Worker 實例。
}

// Download 使用底層 worker 下載指定套件到 package_tmp 環境變數定義的目錄。
//
// 參數:
//   - packageName: 要下載的套件名稱。
//   - indexUrl: 套件的索引 URL。
func (rw WorkerHandler) Download(packageName string, indexUrl string) {
	rw.worker.Download(
		os.Getenv("package_tmp"),
		packageName,
		indexUrl,
	)
}

// DownloadFromIndex 使用底層 worker 從指定索引 URL 下載套件到目的地。
//
// 參數:
//   - destination: 下載套件的目標目錄。
//   - packageName: 要下載的套件名稱。
//   - indexUrl: 套件的索引 URL。
func (rw WorkerHandler) DownloadFromIndex(destination string, packageName string, indexUrl string) {
	rw.worker.Download(
		destination,
		packageName,
		indexUrl,
	)
}

// Sync 使用底層 worker 將指定的套件檔案同步到目標 URL。
//
// 參數:
//   - targetUrl: 目標 URL。
//   - packageFile: 要同步的套件檔案路徑。
func (rw WorkerHandler) Sync(targetUrl string, packageFile string) {
	rw.worker.Sync(
		targetUrl,
		packageFile,
	)
}

// SyncPackagesFromDefintionFile 使用底層 worker 根據定義檔案同步套件。
//
// 參數:
//   - projectName: 專案名稱。
//   - requirementsFile: 包含套件定義的檔案路徑。
func (rw WorkerHandler) SyncPackagesFromDefintionFile(projectName string, requirementsFile string) {
	rw.worker.SyncPackages(projectName, requirementsFile)
}

// Remove 使用底層 worker 刪除指定路徑的套件。
//
// 參數:
//   - fullPath: 要刪除的套件的完整路徑。
//
// 返回:
//   - error: 如果刪除失敗，返回錯誤；否則返回 nil。
func (rw WorkerHandler) Remove(fullPath string) error {
	return rw.worker.Remove(fullPath)
}

// NewRepositoryWorker 創建一個新的 WorkerHandler 實例。
//
// 參數:
//   - worker: 實現 Worker 介面的實例。
//
// 返回:
//   - WorkerHandler: 新創建的 WorkerHandler 實例。
func NewRepositoryWorker(worker Worker) WorkerHandler {
	return WorkerHandler{worker: worker}
}

// UploadToRepository 並行地將源路徑下的所有檔案上傳到目標 URL。
// 它會遍歷源路徑下的所有檔案，為每個檔案啟動一個 goroutine 進行同步上傳。
//
// 參數:
//   - worker: 用於執行同步操作的 WorkerHandler 實例。
//   - targetUrl: 上傳的目標 URL。
//   - sourcePath: 包含要上傳檔案的源目錄路徑。
func UploadToRepository(worker WorkerHandler, targetUrl string, sourcePath string) {

	var wg sync.WaitGroup

	// files, err := ioutil.ReadDir(sourcePath)
	files, err := os.ReadDir(sourcePath)
	if err != nil {
		fmt.Println("讀取目錄錯誤: ", err)
	}
	for _, file := range files {
		wg.Add(1)
		pkgName := fmt.Sprintf("%s/%s", sourcePath, file.Name())
		go func() {
			worker.Sync(targetUrl, pkgName)
			wg.Done()
		}()
	}
	wg.Wait()
}

// GetDependenciesTree 執行命令以獲取依賴樹，並將結果寫入指定檔案。
//
// 參數:
//   - filename: 輸出檔案的名稱，依賴樹將寫入此檔案。
//   - prefix: 命令的前綴 (例如 "gradle" 或 "mvn")。
//   - cmds: 要執行的命令參數。
//
// 返回:
//   - error: 如果命令執行失敗或寫入檔案失敗，返回錯誤；否則返回 nil。
func GetDependenciesTree(filename string, prefix string, cmds []string) error {

	var err error = nil

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, prefix, cmds...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gradle dependencies failed: %w, output %s", err, string(out))
	}
	fmt.Println(string(out))

	err = os.WriteFile(filename, out, 0644) // 0644 是檔案權限，可根據需要調整
	if err != nil {
		return fmt.Errorf("failed to write output to file %s: %w", filename, err)
	}

	return err
}
