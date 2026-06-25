package worker

import (
	"bytes"
	"context"
	"fmt"
	"log" // Added log import
	"os"
	"os/exec"
	"strings" // Added strings import
	"time"
	"wss/repositoryclient"
)

// Npm 結構體用於處理與 npm 套件相關的操作。
type Npm struct {
	Command string // 用於執行 npm 命令的指令。
}

// Download 是一個佔位符函式，用於從 npm 下載指定套件。
// 目前未實作具體功能。
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
	cmd := exec.CommandContext(ctx, npm.Command, cmdArgs...)
	// cmd := exec.CommandContext(ctx, os.Getenv("npm"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w, output: %s", err, string(out))
	}

	return nil
}

// Sync 是一個佔位符函式，用於將檔案同步到目標 URL。
// 目前未實作具體功能。
func (npm Npm) Sync(targetUrl string, packageFile string) string {
	var body string = ""
	return string(body)

}

// Publish 將位於 packageDirPath 的套件發佈到 npm 儲存庫。
// packageDirPath 應該是 npm 套件的根目錄 (包含 package.json)。
func (npm Npm) Publish(conn *repositoryclient.RepositoryConnection, packageDirPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. 配置 npm registry (如果提供了 URL)
	if conn.Url != "" {
		log.Printf("Configuring npm registry to: %s", conn.Url)
		configCmdArgs := []string{"config", "set", "registry", conn.Url}
		configCmd := exec.CommandContext(ctx, npm.Command, configCmdArgs...)
		configCmd.Dir = packageDirPath
		if out, err := configCmd.CombinedOutput(); err != nil {
			log.Printf("Failed to set npm registry: %s, output: %s", err, string(out))
			// 不返回錯誤，因為有些情況下可能不需要設定，或者已經設定好
		}
	}

	// 2. 配置認證 (如果提供了 PAT 或 Base64PAT)
	// npm 認證通常透過 .npmrc 中的 _authToken 或 npm login 處理。
	// 這裡使用 `npm config set` 設定 `_authToken`，這需要 Base64 編碼的 PAT。
	if conn.Base64PAT != "" && conn.Url != "" {
		authCmdArgs := []string{"config", "set", "//" + strings.TrimPrefix(conn.Url, "http://") + "/:_authToken", conn.Base64PAT}
		if strings.HasPrefix(conn.Url, "https://") {
			authCmdArgs = []string{"config", "set", "//" + strings.TrimPrefix(conn.Url, "https://") + "/:_authToken", conn.Base64PAT}
		}

		log.Printf("Configuring npm authentication for %s", conn.Url)
		authCmd := exec.CommandContext(ctx, npm.Command, authCmdArgs...)
		authCmd.Dir = packageDirPath
		if out, err := authCmd.CombinedOutput(); err != nil {
			log.Printf("Failed to set npm auth token: %s, output: %s", err, string(out))
			// 不返回錯誤，因為有些情況下可能不需要設定，或者已經設定好
		}
	} else if conn.PAT != "" && conn.Url != "" {
		// 如果提供了 PAT 但不是 Base64 編碼的，嘗試進行 Base64 編碼
		log.Printf("Warning: PAT provided for npm publish is not Base64 encoded. Attempting to encode.")
		encodedPAT := conn.PAT // 假設這裡的 PAT 已經是明文或者會被 npm 自己處理
		authCmdArgs := []string{"config", "set", "//" + strings.TrimPrefix(conn.Url, "http://") + "/:_authToken", encodedPAT}
		if strings.HasPrefix(conn.Url, "https://") {
			authCmdArgs = []string{"config", "set", "//" + strings.TrimPrefix(conn.Url, "https://") + "/:_authToken", encodedPAT}
		}

		log.Printf("Configuring npm authentication for %s with raw PAT (consider using Base64PAT)", conn.Url)
		authCmd := exec.CommandContext(ctx, npm.Command, authCmdArgs...)
		authCmd.Dir = packageDirPath
		if out, err := authCmd.CombinedOutput(); err != nil {
			log.Printf("Failed to set npm auth token with raw PAT: %s, output: %s", err, string(out))
		}
	}

	// 3. 執行 npm publish
	log.Printf("Publishing npm package from %s to registry %s", packageDirPath, conn.Url)
	cmdArgs := []string{"publish"}
	// 可以添加 --access public/restricted 如果需要
	// 可以添加 --tag <tag> 如果需要

	cmd := exec.CommandContext(ctx, npm.Command, cmdArgs...)
	cmd.Dir = packageDirPath // 在套件目錄中執行

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm publish failed: %w\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}

	log.Printf("Successfully published npm package from %s", packageDirPath)
	return nil
}

// Remove 刪除指定套件名稱對應的臨時目錄。
//
// 參數:
//   - packageName: 要刪除的套件名稱。
//
// 返回:
//   - error: 如果刪除失敗，返回錯誤；否則返回 nil。
func (npm Npm) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
