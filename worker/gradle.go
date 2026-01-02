package worker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Gradle struct{}

type ReplaceRule struct {
	Old string
	New string
}

// FIXME. 移至檔案操作模組
func copyFile(sourcePath, destinationDir string) error {
	// 1. 檢查源檔案是否存在
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("無法開啟源檔案 '%s': %w", sourcePath, err)
	}
	defer sourceFile.Close()

	// 2. 獲取源檔案的檔案資訊 (用於獲取檔案名稱)
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("無法獲取源檔案資訊 '%s': %w", sourcePath, err)
	}
	fileName := fileInfo.Name()

	// 3. 創建目標檔案的路徑
	destinationPath := filepath.Join(destinationDir, fileName)

	// 4. 創建目標檔案
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("無法創建目標檔案 '%s': %w", destinationPath, err)
	}
	defer destinationFile.Close()

	// 5. 複製檔案內容
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("複製檔案內容錯誤: %w", err)
	}

	// 6. 複製檔案權限 (如果需要的話)
	err = os.Chmod(destinationPath, fileInfo.Mode())
	if err != nil {
		fmt.Printf("警告: 無法設置目標檔案權限 '%s': %v\n", destinationPath, err)  // 不返回錯誤，繼續執行
	}

	return nil
}

func readFileContent(filePath string, rules []ReplaceRule) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var content strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}
		for _, rule := range rules {
			line = strings.ReplaceAll(line, rule.Old, rule.New)
		}

		content.WriteString(line)
	}
	return content.String(), nil
}

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
func (gradle Gradle) SyncPackages(destination string, requirementsFile string) error {

	var err error = nil
	packageTmp := os.Getenv("package_tmp")
	tplFile := "./templates/build_tasks.gradle"
	downloadDestination := fmt.Sprintf("%s/%s", packageTmp, destination)
	replaceRules := []ReplaceRule{
		{Old: "<destPath>", New: downloadDestination},
	}

	tplContent, err := readFileContent(tplFile, replaceRules)
	if err != nil {
		fmt.Println("Error reading tplFile:", err)
		return err
	}
	err = appendToFile(requirementsFile, tplContent)
	if err != nil {
		fmt.Println("Error appending to requirementsFile:", err)
		return err
	}
	if err := os.MkdirAll(downloadDestination, 0755); err != nil {
		return fmt.Errorf("Create Dir failed:%w", err)
	}
	// gradle -p /Users/ccxn/Desktop/demoGradle/ downloadDependencies

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmdArgs := []string{"-p", "./", "downloadDependencies"}
	cmd := exec.CommandContext(ctx, os.Getenv("gradle"), cmdArgs...)
	out, err := cmd.CombinedOutput()
	fmt.Println("out: ", string(out))
	if err != nil {
		return fmt.Errorf("gradle download failed: %w, output %s", err, string(out))
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
