package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Mvn struct{}

func (mvn Mvn) Download(destination string, packageName string, indexUrl string) string {
	var out string = ""
	return string(out)
}

func (mvn Mvn) SyncPackages(destination string, requirementsFile string) error {
	// 1. 驗證環境變數
	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	reportTmp := os.Getenv("report_tmp")
	if packageTmp == "" {
		return fmt.Errorf("environment variable 'package_tmp' is not set")
	}

	// 2. 驗證 pom.xml 是否存在 (提早失敗，避免執行無意義的 Maven 指令)
	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		return fmt.Errorf("requirements file not found: %s", requirementsFile)
	}

	// 3. 建構路徑 (使用 filepath.Join 確保跨平台相容性)
	// downloadDestination: 這是你要存放乾淨 jar 檔的地方 (給掃描器用)
	downloadDestination := filepath.Join(packageTmp, destination)
	reportDestination := filepath.Join(reportTmp, destination)

	// localRepo: 這是 Maven 的暫存倉庫 (結構複雜，含 pom/metadata)
	// 建議與 downloadDestination 分開，避免掃描器掃到不該掃的 cache 檔
	// localRepo := filepath.Join(packageTmp, destination+"_m2_repo")

	// 建立目錄
	if err = os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("failed to create download dir: %w", err)
	}

	if err = os.MkdirAll(reportDestination, 0755); err != nil {
		return fmt.Errorf("failed to create download dir: %w", err)
	} else {
		fmt.Printf("create report directory %s successful.", reportDestination)
	}
	// localRepo 可以讓 Maven 自己建，但為了保險也可以先建
	// if err := os.MkdirAll(localRepo, 0755); err != nil {
	// 	return fmt.Errorf("failed to create local repo dir: %w", err)
	// }

	// 4. 準備 Maven 指令參數
	cmdArgs := []string{
		// 強制非互動模式，CI/CD 環境必備
		"-B",
		// 使用 dependency:copy-dependencies 插件
		"dependency:copy-dependencies",
		// 指定 pom 檔案路徑
		"-f", requirementsFile,
		// 指定本地倉庫緩存 (加速後續執行，且隔離環境)
		fmt.Sprintf("-Dmaven.repo.local=%s", downloadDestination),
		// 指定「乾淨 jar 檔」的輸出位置 (這才是 copy-dependencies 的目標)
		// fmt.Sprintf("-DoutputDirectory=%s", downloadDestination),
		// 建議：只下載 runtime 依賴 (排除 test)，除非你要掃描測試程式碼
		"-DincludeScope=runtime",
		// 建議：若有 snapshot 版本則強制更新
		"-U",
	}

	// Debug 用：印出完整指令方便除錯
	fmt.Printf("Executing Maven: mvn %v\n", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // 設定 10 分鐘超時
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Getenv("mvn"), cmdArgs...)

	// 5. 執行並捕捉輸出
	// 使用 CombinedOutput 可以在錯誤時一次把 stdout/stderr 印出來
	out, err := cmd.CombinedOutput()
	if err != nil {
		// 回傳錯誤時，把 Maven 的 log 附帶在 error message 裡，方便追查
		// 注意：如果 log 太長可能要考慮截斷，但在這裡全印出來比較保險
		return fmt.Errorf("mvn execution failed: %w\nOutput:\n%s", err, string(out))
	}

	return nil
}

// func (mvn Mvn) SyncPackages(destination string, requirementsFile string) error {
// 	packageTmp := os.Getenv("package_tmp")
// 	if packageTmp == "" {
// 		return fmt.Errorf("package_tmp is empty")
//
// 	}
//
// 	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
// 	if err := os.MkdirAll(downloadDestination, 0755); err != nil {
// 		return fmt.Errorf("Create Dir failed: %w", err)
// 	}
// 	outputDirectory := fmt.Sprintf("-Dmaven.repo.local=%s", downloadDestination)
// 	cmdArgs := []string{"dependency:copy-dependencies", "-f", requirementsFile, outputDirectory}
// 	fmt.Println(cmdArgs)
// 	cmd := exec.Command("mvn", cmdArgs...)
// 	out, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("mvn download failed: %w, output: %s", err, string(out))
// 	}
// 	return nil
// }

func (mvn Mvn) Sync(targetUrl string, packageFile string) string {

	var body string = ""
	return string(body)

}

func (mvn Mvn) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
