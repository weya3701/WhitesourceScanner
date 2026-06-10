package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Gradle struct {
	Command string
}

type ReplaceRule struct {
	Old string
	New string
}

// FIXME. 移至檔案操作模組
// func copyFile(sourcePath, destinationDir string) error {
// 	// 1. 檢查源檔案是否存在
// 	sourceFile, err := os.Open(sourcePath)
// 	if err != nil {
// 		return fmt.Errorf("無法開啟源檔案 '%s': %w", sourcePath, err)
// 	}
// 	defer sourceFile.Close()
//
// 	// 2. 獲取源檔案的檔案資訊 (用於獲取檔案名稱)
// 	fileInfo, err := sourceFile.Stat()
// 	if err != nil {
// 		return fmt.Errorf("無法獲取源檔案資訊 '%s': %w", sourcePath, err)
// 	}
// 	fileName := fileInfo.Name()
//
// 	// 3. 創建目標檔案的路徑
// 	destinationPath := filepath.Join(destinationDir, fileName)
//
// 	// 4. 創建目標檔案
// 	destinationFile, err := os.Create(destinationPath)
// 	if err != nil {
// 		return fmt.Errorf("無法創建目標檔案 '%s': %w", destinationPath, err)
// 	}
// 	defer destinationFile.Close()
//
// 	// 5. 複製檔案內容
// 	_, err = io.Copy(destinationFile, sourceFile)
// 	if err != nil {
// 		return fmt.Errorf("複製檔案內容錯誤: %w", err)
// 	}
//
// 	// 6. 複製檔案權限 (如果需要的話)
// 	err = os.Chmod(destinationPath, fileInfo.Mode())
// 	if err != nil {
// 		fmt.Printf("警告: 無法設置目標檔案權限 '%s': %v\n", destinationPath, err)  // 不返回錯誤，繼續執行
// 	}
//
// 	return nil
// }

func getBuildTemplate() string {
	return `task downloadDependencies(type: Copy) {
		from configurations.runtimeClasspath
		into "%s"
	}
	`
}

// readFileContent 函數從指定文件路徑讀取文件內容，並根據提供的替換規則進行文本替換。
//
// 參數：
//   - filePath: 要讀取的文件路徑。
//   - rules: 替換規則的切片。每個 ReplaceRule 包含要查找的舊字符串和要替換的新字符串。
//
// 返回值：
//   - string: 處理後的文件的內容。如果發生錯誤，則返回空字符串。
//   - error: 如果發生錯誤，則返回錯誤信息；否則返回 nil。
// func readFileContent(filePath string, rules []ReplaceRule) (string, error) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer file.Close()
//
// 	reader := bufio.NewReader(file)
// 	var content strings.Builder
//
// 	for {
// 		line, err := reader.ReadString('\n')
// 		if err == io.EOF {
// 			break
// 		}
//
// 		if err != nil {
// 			return "", err
// 		}
// 		for _, rule := range rules {
// 			line = strings.ReplaceAll(line, rule.Old, rule.New)
// 		}
//
// 		content.WriteString(line)
// 	}
// 	return content.String(), nil
// }

// File 將內容追加到指定的文件中。
//
// 參數：
//   - filePath: 要追加內容的文件路徑。
//   - content: 要追加到文件的字符串內容。
//
// 傳回值：
//   - error
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

func (gradle Gradle) Download(destination string, packageName string, indexUrl string) string {
	var cmd string = "Need to implement"
	return string(cmd)
}

// FIXME. 需新增./templates/build_tasks.gradled檔案
// SyncPackages 函數用於同步 Gradle 專案的套件依賴。
// 該函數通過以下步驟實現：
//  1. 從模板檔案 `./templates/build_tasks.gradle` 讀取內容，該檔案包含用於下載依賴項的 Gradle 任務。
//  2. 替換模板中的特定佔位符，例如將 `<destPath>` 替換為下載目標路徑。
//  3. 將修改後的 Gradle 任務內容追加到需求檔案 `requirementsFile` 中。
//  4. 建立用於儲存下載套件的目錄。
//  5. 執行 Gradle 命令，使用 `-p .` 指定專案根目錄，並執行 `downloadDependencies` 任務來下載依賴項。
//  6. 建立用於儲存報告的目錄。
//  7. 執行 Gradle 命令，獲取依賴樹。
//
// 參數:
//   - destination:  套件的目標目錄，通常是專案名稱或版本。
//   - requirementsFile:  用於存儲 Gradle 配置的檔案路徑，會將 downloadDependencies 任务 添加到该文件
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

func (gradle Gradle) Sync(targetUrl string, packageFile string) string {
	var output string = "Need to implement"
	return string(output)

}

func (gradle Gradle) Remove(packageName string) error {
	fullPath := fmt.Sprintf("./tmp/%s", packageName)
	err := os.RemoveAll(fullPath)
	return err

}
