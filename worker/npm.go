package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Npm struct{}

func (npm Npm) Download(destination string, packageName string, indexUrl string) string {
	var cmd string = ""
	return string(cmd)
}

// SyncPackages 函數同步 npm 套件。
//
// 參數：
//   - destination:  要同步套件的目的地目錄名稱 (例如 "package1")。此名稱將用作子目錄，在 package_tmp 和 report_tmp 目錄中創建。
//   - requirementsFile:  package.json 檔案的完整路徑，該檔案包含要安裝的套件。  此檔案將被複製到下載目錄中。
//
// 返回值：
//   - error:  如果發生任何錯誤，將返回一個錯誤。否則，返回 nil 表示成功。
//
// 流程：
//  1. 獲取 "package_tmp" 和 "report_tmp" 環境變數的值。
//  2. 檢查 "package_tmp" 是否為空。如果是，則返回錯誤。
//  3. 根據 "package_tmp"、"report_tmp" 和提供的目的地創建下載和報告目錄的完整路徑。
//  4. 使用 os.MkdirAll 創建下載和報告目錄。如果目錄已存在，則不會返回錯誤。
//  5. 將 requirementsFile (package.json) 檔案複製到下載目錄。
//  6. 使用 `npm install --prefix <downloadDestination>` 命令，在下載目錄中安裝套件。
//  7. 如果在任何步驟中發生錯誤，則返回錯誤。
func (npm Npm) SyncPackages(destination string, requirementsFile string) error {

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
	copyCmd := []string{requirementsFile, downloadDestination}
	cpcmd := exec.Command("cp", copyCmd...)
	cpout, err := cpcmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Copy package.json failed: %w, output: %s", err, string(cpout))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmdArgs := []string{"install", "-prefix", downloadDestination}
	cmd := exec.CommandContext(ctx, os.Getenv("npm"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pip download failed: %w, output: %s", err, string(out))
	}

	return nil
}

func (npm Npm) Sync(targetUrl string, packageFile string) string {
	var body string = ""
	return string(body)

}

func (npm Npm) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
