package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Gradle 結構體用於處理與 Gradle 相關的操作。
type Gradle struct {
	Command string // 用於執行 Gradle 命令的指令。
}

// ReplaceRule 結構體定義了一個替換規則，包含舊字符串和新字符串。
type ReplaceRule struct {
	Old string // 要搜尋的舊字符串。
	New string // 替換舊字符串的新字符串。
}

// getBuildTemplate 返回一個 Gradle 任務的字串模板，用於下載依賴項。
// 該模板包含一個 `downloadDependencies` 任務，會將運行時依賴複製到指定目錄。
//
// 返回:
//   - string: Gradle 任務的字串模板。
func getBuildTemplate() string {
	return `task downloadDependencies(type: Copy) {
		from configurations.runtimeClasspath
		into "%s"
	}
	`
}

// appendToFile 將內容追加到指定的文件中。
//
// 參數：
//   - filePath: 要追加內容的文件路徑。
//   - content: 要追加到文件的字符串內容。
//
// 傳回值：
//   - error: 如果開啟文件或寫入內容失敗，返回錯誤；否則返回 nil。
func appendToFile(filePath string, content string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

// Download 是一個佔位符函式，用於從 Gradle 倉庫下載指定套件。
// 目前未實作具體功能。
func (gradle Gradle) Download(destination string, packageName string, indexUrl string) string {
	var cmd string = "Need to implement"
	return string(cmd)
}

// SyncPackages 函數用於同步 Gradle 專案的套件依賴。
// 該函數通過以下步驟實現：
//  1. 根據下載目的地格式化 Gradle 任務內容。
//  2. 將修改後的 Gradle 任務內容追加到需求檔案 `requirementsFile` 中。
//  3. 建立用於儲存下載套件的目錄。
//  4. 建立用於儲存報告的目錄。
//  5. 執行 Gradle 命令，使用 `-p .` 指定專案根目錄，並執行 `downloadDependencies` 任務來下載依賴項。
//  6. 執行 Gradle 命令，獲取依賴樹並儲存到檔案。
//
// 參數:
//   - destination:  套件的目標目錄，通常是專案名稱或版本。
//   - requirementsFile:  用於存儲 Gradle 配置的檔案路徑，會將 downloadDependencies 任务 添加到該文件
//
// 返回:
//   - error: 如果在任何步驟中發生錯誤，則返回錯誤訊息；否則返回 nil。
func (gradle Gradle) SyncPackages(destination string, requirementsFile string) error {
	// 取得./templates/build_tasks.gradle將downloadDependencies功能加入build.gradle
	// 再進行gradle -p ./ downloadDependencies命令取得套件。

	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	reportTmp := os.Getenv("report_tmp")
	// tplFile := "./templates/build_tasks.gradle"
	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
	reportDestination := fmt.Sprintf("%s/%s", reportTmp, destination)
	// replaceRules := []ReplaceRule{
	// 	{Old: "<destPath>", New: downloadDestination},
	// }

	// tplContent, err := readFileContent(getBuildTemplate(), replaceRules)
	tplContent := fmt.Sprintf(getBuildTemplate(), downloadDestination)
	// if err != nil {
	// 	fmt.Println("Error reading tplFile:", err)
	// 	return err
	// }
	err = appendToFile(requirementsFile, tplContent)
	if err != nil {
		fmt.Println("Error appending to requirementsFile:", err)
		return err
	}
	if err := os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("Create pakcages directory failed:%w", err)
	}
	if err = os.MkdirAll(reportDestination, 0755); err != nil {
		return fmt.Errorf("Create report directory failed: %w", err)
	} else {
		fmt.Printf("Create %s successful.", reportDestination)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmdArgs := []string{"-p", "./", "downloadDependencies"}
	cmd := exec.CommandContext(ctx, gradle.Command, cmdArgs...)
	// cmd := exec.CommandContext(ctx, os.Getenv("gradle"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gradle download failed: %w, output %s", err, string(out))
	}
	dependenciesTreeFile := fmt.Sprintf("%s/dependenciesTree.txt", reportDestination)
	fmt.Println("dependencies tree file: ", dependenciesTreeFile)
	cmds := []string{"dependencies"}
	err = GetDependenciesTree(dependenciesTreeFile, gradle.Command, cmds)

	if err != nil {
		return fmt.Errorf("Get dependencies tree failed: %w", err)
	}
	return nil
}

// Sync 是一個佔位符函式，用於將檔案同步到目標 URL。
// 目前未實作具體功能。
func (gradle Gradle) Sync(targetUrl string, packageFile string) string {
	var output string = "Need to implement"
	return string(output)

}

// Remove 刪除指定套件名稱對應的臨時目錄。
//
// 參數:
//   - packageName: 要刪除的套件名稱。
//
// 返回:
//   - error: 如果刪除失敗，返回錯誤；否則返回 nil。
func (gradle Gradle) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
