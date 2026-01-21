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

// SyncPackages 函數用於同步 Maven 依賴項到指定目的地。
//
// 它執行以下步驟：
// 1. 獲取環境變數 `$package_tmp` 和 `$report_tmp`，用於儲存下載的包和報告。
// 2. 檢查 `$package_tmp` 是否已設定。
// 3. 檢查 `requirementsFile` 文件是否存在。
// 4. 根據 `$package_tmp` 和 `destination` 建立下載目錄。
// 5. 根據 `$report_tmp` 和 `destination` 建立報告目錄。
// 6. 使用 Maven 命令 `dependency:copy-dependencies` 從 `requirementsFile` 下載依賴項到下載目錄。
// 7. 使用 `dependency:tree` 命令生成依賴項樹並將其儲存到文件。
//
// 參數：
//   - mvn: 指向 Mvn 結構體的指標 (未使用，可能為預留)。
//   - destination: 字符串，用於指定下載和報告的相對路徑。  這個路徑會被加入到 `$package_tmp` 和 `$report_tmp` 環境變數指定的基礎路徑中，形成完整的下載和報告路徑。
//   - requirementsFile: 字符串，指向包含 Maven 依賴項的文件。
//
// 返回值：
//   - error: 如果發生錯誤，則返回錯誤；否則返回 nil。
func (mvn Mvn) SyncPackages(destination string, requirementsFile string) error {
	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	reportTmp := os.Getenv("report_tmp")
	if packageTmp == "" {
		return fmt.Errorf("environment variable 'package_tmp' is not set")
	}

	if _, err := os.Stat(requirementsFile); os.IsNotExist(err) {
		return fmt.Errorf("requirements file not found: %s", requirementsFile)
	}

	downloadDestination := filepath.Join(packageTmp, destination)
	reportDestination := filepath.Join(reportTmp, destination)

	if err = os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("failed to create download dir: %w", err)
	}

	if err = os.MkdirAll(reportDestination, 0755); err != nil {
		return fmt.Errorf("failed to create download dir: %w", err)
	} else {
		fmt.Printf("create report directory %s successful.", reportDestination)
	}

	cmdArgs := []string{
		"-B",
		"dependency:copy-dependencies",
		"-f", requirementsFile,
		fmt.Sprintf("-Dmaven.repo.local=%s", downloadDestination),
		"-DincludeScope=runtime",
		"-U",
	}

	fmt.Printf("Executing Maven: mvn %v\n", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // 設定 10 分鐘超時
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Getenv("mvn"), cmdArgs...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mvn execution failed: %w\nOutput:\n%s", err, string(out))
	}

	dependenciesTreeFile := fmt.Sprintf("%s/dependenciesTree.txt", reportDestination)
	cmds := []string{"dependency:tree", "-f", requirementsFile, fmt.Sprintf("-Dmaven.repo.local=%s", downloadDestination), "-DincludeScope=runtime", "-U"}
	GetDependenciesTree(dependenciesTreeFile, os.Getenv("mvn"), cmds)

	return nil
}

func (mvn Mvn) Sync(targetUrl string, packageFile string) string {

	var body string = ""
	return string(body)

}

func (mvn Mvn) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
