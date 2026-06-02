package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestMainModes 測試 main 函數在不同模式下的行為。
// 注意：這個測試會觸發對 handler 和 wss 函數的實際調用，
// 這可能涉及網路請求、檔案系統操作或外部命令執行。
// 為了進行真正的單元測試，其依賴項（handler, wss）需要被模擬 (mock)。
func TestMainModes(t *testing.T) {
	// 1. 保存並恢復原始的 os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// 2. 保存並恢復原始的 stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() { os.Stdout = oldStdout; os.Stderr = oldStderr }()

	// 捕獲輸出
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// 創建一個臨時目錄，用於測試特有的檔案 (例如：.env, report 目錄, whitesource_path 目錄)
	tempDir, err := ioutil.TempDir("", "main_test_")
	if err != nil {
		t.Fatalf("創建臨時目錄失敗: %v", err)
	}
	defer os.RemoveAll(tempDir) // 清理臨時目錄

	// 暫時將當前工作目錄更改為臨時目錄，以便 godotenv.Load(".env") 能找到
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd) // 恢復原始工作目錄

	// 3. 創建一個空的 .env 檔案，供 godotenv.Load() 使用
	dotEnvPath := filepath.Join(tempDir, ".env") // 在臨時目錄中
	err = ioutil.WriteFile(dotEnvPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("創建假的 .env 檔案失敗: %v", err)
	}

	// 4. 創建一個假的 settings_file (YAML 格式)，供 wss.DoWhitesourceScan 使用
	settingsFilePath := filepath.Join(tempDir, "test_settings.yaml")
	settingsContent := `
apiKey: "test_api_key"
userKey: "test_user_key"
projectName: "test_project"
productName: "test_product"
productToken: "test_product_token"
wss.url: "http://example.com/api"
offline: "true"
`
	err = ioutil.WriteFile(settingsFilePath, []byte(settingsContent), 0644)
	if err != nil {
		t.Fatalf("創建假的 settings_file 失敗: %v", err)
	}

	// 5. 設置 handler/wss 函數所依賴的環境變數
	os.Setenv("settings_file", settingsFilePath)
	// whitesource_path 需要是存在的目錄，以便 GetFilePath 不會失敗
	os.Setenv("whitesource_path", filepath.Join(tempDir, "whitesource_data"))
	os.Setenv("request_file", "request.json") // 由 GetProjectAlert 和 GetProjectRiskReport 使用

	defer os.Unsetenv("settings_file")
	defer os.Unsetenv("whitesource_path")
	defer os.Unsetenv("request_file")

	// 創建 whitesource_path 目錄，因為 GetFilePath 期望它存在
	os.MkdirAll(filepath.Join(tempDir, "whitesource_data"), 0755)

	// 定義測試案例
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "mode_reqfile",
			args: []string{"main", "-mode", "reqfile", "-package_type", "python", "-project_name", "test-project-reqfile", "-requirements_file", "requirements.txt", "-package_name", "test-pkg", "-with_conf", "true"},
		},
		{
			name: "mode_cmd",
			args: []string{"main", "-mode", "cmd", "-package_name", "test-package", "-project_name", "test-project-cmd", "-with_conf", "false", "-package_type", "python"},
		},
		{
			name: "mode_image",
			args: []string{"main", "-mode", "image", "-export_file", "export.json", "-application", "test-app", "-package_name", "test-image-pkg", "-project_name", "test-project-image", "-tar_file", "image.tar", "-image_name", "myimage", "-image_tag", "latest"},
		},
		{
			name: "mode_empty", // 測試沒有有效模式的情況
			args: []string{"main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 為當前測試運行重置 os.Args
			os.Args = tt.args
			// 重置 flag 套件的狀態，因為 flag.Parse() 在 main 函數中被調用
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// 執行 main 函數
			// 使用 goroutine 和 recover 來捕獲 main() 可能拋出的任何 panic，
			// 這可以防止整個測試套件因一個 panic 而中止。
			done := make(chan bool)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("main() 發生 panic: %v", r)
					}
					done <- true
				}()
				main()
			}()
			<-done // 等待 main goroutine 執行完成

			// 關閉 stdout/stderr 的寫入管道，以便從讀取管道中獲取輸出
			wOut.Close()
			wErr.Close()

			var bufOut, bufErr bytes.Buffer
			_, _ = bufOut.ReadFrom(rOut)
			_, _ = bufErr.ReadFrom(rErr)

			if bufErr.Len() > 0 {
				t.Logf("模式 %s 的標準錯誤輸出:\n%s", tt.name, bufErr.String())
				// 由於外部依賴項未被模擬，某些 handler/wss 函數可能會打印錯誤到 stderr，
				// 但 main 函數本身沒有 panic。這裡僅記錄不作為測試失敗條件。
			}
			if bufOut.Len() > 0 {
				t.Logf("模式 %s 的標準輸出:\n%s", tt.name, bufOut.String())
			}

			// 為下一個測試運行重新打開管道
			rOut, wOut, _ = os.Pipe()
			rErr, wErr, _ = os.Pipe()
			os.Stdout = wOut
			os.Stderr = wErr
		})
	}
}
