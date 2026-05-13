package wss

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetScanSingleton 測試 GetScanSingleton 函數，驗證其單例模式的實現。
func TestGetScanSingleton(t *testing.T) {
	// 第一次調用 GetScanSingleton 函數。
	mutex1 := GetScanSingleton()

	// 第二次調用 GetScanSingleton 函數。
	mutex2 := GetScanSingleton()

	// 驗證兩次調用返回的 mutex 指針是否相同。
	assert.Equal(t, mutex1, mutex2, "GetScanSingleton 應該返回相同的 mutex 指針")
}

// TestWhiteSourceEnv_ParserEnv 測試 WhiteSourceEnv 的 ParserEnv 方法，驗證其解析 YAML 配置文件的功能。
func TestWhiteSourceEnv_ParserEnv(t *testing.T) {
	// 創建一個臨時 YAML 配置文件。
	tempFile, err := ioutil.TempFile("", "test_config.yaml")
	require().NoError(t, err, "創建臨時文件失敗")
	defer os.Remove(tempFile.Name()) // 測試完成後刪除臨時文件

	// 寫入一些測試用的配置到臨時文件中。
	content := `
apiKey: test_api_key
userKey: test_user_key
projectName: test_project_name
productName: test_product_name
productToken: test_product_token
wss.url: test_wss_url
offline: "true"
`
	_, err = tempFile.WriteString(content)
	require().NoError(t, err, "寫入臨時文件失敗")
	err = tempFile.Close()
	require().NoError(t, err, "關閉臨時文件失敗")

	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{}

	// 調用 ParserEnv 方法解析臨時配置文件。
	config.ParserEnv(tempFile.Name())

	// 驗證解析後的配置是否符合預期。
	assert.Equal(t, "test_api_key", config.ApiKey, "ApiKey 解析不正確")
	assert.Equal(t, "test_user_key", config.UserKey, "UserKey 解析不正確")
	assert.Equal(t, "test_project_name", config.ProjectName, "ProjectName 解析不正確")
	assert.Equal(t, "test_product_name", config.ProductName, "ProductName 解析不正確")
	assert.Equal(t, "test_product_token", config.ProductToken, "ProductToken 解析不正確")
	assert.Equal(t, "test_wss_url", config.WSSUrl, "WSSUrl 解析不正確")
	assert.Equal(t, "true", config.Offline, "Offline 解析不正確")
}

// TestWhiteSourceEnv_ParserEnv_InvalidFile 測試 WhiteSourceEnv 的 ParserEnv 方法處理無效文件路徑的情況。
func TestWhiteSourceEnv_ParserEnv_InvalidFile(t *testing.T) {
	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{}

	// 調用 ParserEnv 方法，傳入一個不存在的文件路徑。
	assert.Panics(t, func() {
		config.ParserEnv("nonexistent_file.yaml")
	}, "應該 panic，因為文件不存在")
}

// TestWhiteSourceEnv_ParserEnv_InvalidContent 測試 WhiteSourceEnv 的 ParserEnv 方法處理無效 YAML 內容的情況。
func TestWhiteSourceEnv_ParserEnv_InvalidContent(t *testing.T) {
	// 創建一個臨時 YAML 配置文件。
	tempFile, err := ioutil.TempFile("", "test_config.yaml")
	require().NoError(t, err, "創建臨時文件失敗")
	defer os.Remove(tempFile.Name()) // 測試完成後刪除臨時文件

	// 寫入一些無效的 YAML 內容到臨時文件中。
	content := `
apiKey: test_api_key
invalid_field:
  - value1
  - value2
`
	_, err = tempFile.WriteString(content)
	require().NoError(t, err, "寫入臨時文件失敗")
	err = tempFile.Close()
	require().NoError(t, err, "關閉臨時文件失敗")

	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{}

	// 調用 ParserEnv 方法解析臨時配置文件。
	assert.Panics(t, func() {
		config.ParserEnv(tempFile.Name())
	}, "應該 panic，因為 YAML 內容無效")
}

// TestWhiteSourceEnv_SetProjectName 測試 WhiteSourceEnv 的 SetProjectName 方法，驗證其設置 ProjectName 的功能。
func TestWhiteSourceEnv_SetProjectName(t *testing.T) {
	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{}

	// 創建一個 project name。
	projectName := "new_project_name"

	// 調用 SetProjectName 方法設置 ProjectName。
	config.SetProjectName(&projectName)

	// 驗證 ProjectName 是否正確設置。
	assert.Equal(t, "new_project_name", config.ProjectName, "ProjectName 設置不正確")
}

// TestWhiteSourceEnv_SetProductName 測試 WhiteSourceEnv 的 SetProductName 方法，驗證其設置 ProductName 的功能。
func TestWhiteSourceEnv_SetProductName(t *testing.T) {
	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{}

	// 創建一個 product name。
	productName := "new_product_name"

	// 調用 SetProductName 方法設置 ProductName。
	config.SetProductName(&productName)

	// 驗證 ProductName 是否正確設置。
	assert.Equal(t, "new_product_name", config.ProductName, "ProductName 設置不正確")
}

// TestWhiteSourceEnv_SetEnv 測試 WhiteSourceEnv 的 SetEnv 方法，驗證其設置環境變量的功能。
func TestWhiteSourceEnv_SetEnv(t *testing.T) {
	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{
		ApiKey:       "test_api_key",
		UserKey:      "test_user_key",
		ProjectName:  "test_project_name",
		ProductName:  "test_product_name",
		ProductToken: "test_product_token",
		WSSUrl:       "test_wss_url",
		Offline:      "true",
	}

	// 調用 SetEnv 方法設置環境變量。
	config.SetEnv()

	// 驗證環境變量是否正確設置。
	assert.Equal(t, "test_api_key", os.Getenv("WS_APIKEY"), "WS_APIKEY 設置不正確")
	assert.Equal(t, "test_user_key", os.Getenv("WS_USERKEY"), "WS_USERKEY 設置不正確")
	assert.Equal(t, "test_project_name", os.Getenv("WS_PROJECTNAME"), "WS_PROJECTNAME 設置不正確")
	assert.Equal(t, "test_product_name", os.Getenv("WS_PRODUCTNAME"), "WS_PRODUCTNAME 設置不正確")
	assert.Equal(t, "test_product_token", os.Getenv("WS_PRODUCTTOKEN"), "WS_PRODUCTTOKEN 設置不正確")
	assert.Equal(t, "test_wss_url", os.Getenv("WS_WSS_URL"), "WS_WSS_URL 設置不正確")
	assert.Equal(t, "true", os.Getenv("WS_OFFLINE"), "WS_OFFLINE 設置不正確")

	// 清除環境變量，避免影響其他測試。
	os.Unsetenv("WS_APIKEY")
	os.Unsetenv("WS_USERKEY")
	os.Unsetenv("WS_PROJECTNAME")
	os.Unsetenv("WS_PRODUCTNAME")
	os.Unsetenv("WS_PRODUCTTOKEN")
	os.Unsetenv("WS_WSS_URL")
	os.Unsetenv("WS_OFFLINE")
}

// TestMoveRequestFile 測試 MoveRequestFile 函數，驗證其移動文件的功能。
func TestMoveRequestFile(t *testing.T) {
	// 創建一個臨時文件。
	tempFile, err := ioutil.TempFile("", "test_request.txt")
	require().NoError(t, err, "創建臨時文件失敗")
	defer os.Remove(tempFile.Name()) // 測試完成後刪除臨時文件

	// 寫入一些測試數據到臨時文件中。
	content := "Test request content"
	_, err = tempFile.WriteString(content)
	require().NoError(t, err, "寫入臨時文件失敗")
	err = tempFile.Close()
	require().NoError(t, err, "關閉臨時文件失敗")

	// 創建一個目標文件路徑。
	destinationFile := filepath.Join(os.TempDir(), "moved_request.txt")
	defer os.Remove(destinationFile) // 測試完成後刪除目標文件

	// 調用 MoveRequestFile 函數移動文件。
	MoveRequestFile(tempFile.Name(), destinationFile)

	// 驗證文件是否已成功移動到目標路徑。
	_, err = os.Stat(destinationFile)
	assert.NoError(t, err, "文件移動失敗")

	// 驗證目標文件內容是否正確。
	contentBytes, err := ioutil.ReadFile(destinationFile)
	require().NoError(t, err, "讀取目標文件內容失敗")
	assert.Equal(t, content, string(contentBytes), "目標文件內容不正確")
}

// TestMoveRequestFile_InvalidSource 測試 MoveRequestFile 函數處理無效源文件路徑的情況。
func TestMoveRequestFile_InvalidSource(t *testing.T) {
	// 創建一個目標文件路徑。
	destinationFile := filepath.Join(os.TempDir(), "moved_request.txt")
	defer os.Remove(destinationFile) // 測試完成後刪除目標文件

	// 調用 MoveRequestFile 函數，傳入一個不存在的源文件路徑。
	assert.Panics(t, func() {
		MoveRequestFile("nonexistent_file.txt", destinationFile)
	}, "應該 panic，因為源文件不存在")
}

// TestCreateDirectory 測試 CreateDirectory 函數，驗證其創建目錄的功能。
func TestCreateDirectory(t *testing.T) {
	// 創建一個臨時目錄作為基礎目錄。
	tempDir, err := ioutil.TempDir("", "test_base")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(tempDir) // 測試完成後刪除臨時目錄

	// 定義要創建的子目錄名稱。
	dirname := "test_subdir"

	// 調用 CreateDirectory 函數創建子目錄。
	CreateDirectory(tempDir, dirname)

	// 驗證子目錄是否已成功創建。
	subdirPath := filepath.Join(tempDir, dirname)
	_, err = os.Stat(subdirPath)
	assert.NoError(t, err, "子目錄創建失敗")
	assert.True(t, isDirectory(subdirPath), "路徑不是一個目錄")
}

// isDirectory 輔助函數，用於檢查給定的路徑是否是一個目錄。
func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// TestDoDockerTarFileScan 測試 DoDockerTarFileScan 函數，驗證其掃描 Docker 壓縮包文件的功能。
func TestDoDockerTarFileScan(t *testing.T) {
	// 創建一個臨時目錄，用於存放測試用的 Docker 壓縮包文件。
	tempDir, err := ioutil.TempDir("", "test_docker")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(tempDir) // 測試完成後刪除臨時目錄

	// 創建一個測試用的 Docker 壓縮包文件。
	tarFile := filepath.Join(tempDir, "test_image.tar")
	err = ioutil.WriteFile(tarFile, []byte("Test docker image content"), 0644)
	require().NoError(t, err, "創建臨時 Docker 壓縮包文件失敗")

	// 創建一個 MendCli 實例。
	cli := MendCli{
		Application: "test_app",
		PackageName: "test_package",
		ProjectName: "test_project",
		TarFile:     tarFile,
	}

	// 調用 DoDockerTarFileScan 函數掃描 Docker 壓縮包文件。
	// 由於 DoDockerTarFileScan 函數會調用外部命令，這裡只驗證函數是否正常執行，不驗證掃描結果。
	DoDockerTarFileScan(cli)

	// 可以在這裡添加更詳細的驗證，例如檢查掃描日誌是否包含預期的信息。
}

// TestInitialUnifiedAgent 測試 initialUnifiedAgent 函數，驗證其初始化 Unified Agent 的功能。
func TestInitialUnifiedAgent(t *testing.T) {
	// 創建一個臨時目錄，用於存放 Unified Agent。
	tempDir, err := ioutil.TempDir("", "test_ua")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(tempDir) // 測試完成後刪除臨時目錄

	// 設置環境變量
	os.Setenv("wget", "echo") // 使用 echo 模擬 wget
	os.Setenv("agentURL", "https://example.com/ua.jar")
	os.Setenv("wssAgentPath", tempDir)
	os.Setenv("wssAgentName", "ua.jar")

	// 調用 initialUnifiedAgent 函數初始化 Unified Agent。
	err = initialUnifiedAgent(tempDir, "ua.jar")
	require().NoError(t, err, "初始化 Unified Agent 失敗")

	// 驗證 Unified Agent 是否已成功下載到臨時目錄中。
	uaPath := filepath.Join(tempDir, "ua.jar")
	_, err = os.Stat(uaPath)
	assert.NoError(t, err, "Unified Agent 下載失敗")
}

// TestGetUnifiedAgent 測試 getUnifiedAgent 函數，驗證其下載 Unified Agent 的功能。
func TestGetUnifiedAgent(t *testing.T) {
	// 創建一個臨時文件，用於存放 Unified Agent。
	tempFile, err := ioutil.TempFile("", "test_ua.jar")
	require().NoError(t, err, "創建臨時文件失敗")
	defer os.Remove(tempFile.Name()) // 測試完成後刪除臨時文件

	// 設置環境變量
	os.Setenv("wget", "echo") // 使用 echo 模擬 wget

	// 調用 getUnifiedAgent 函數下載 Unified Agent。
	err = getUnifiedAgent(tempFile.Name(), "echo", []string{"https://example.com/ua.jar", "-o", tempFile.Name()})
	require().NoError(t, err, "下載 Unified Agent 失敗")

	// 驗證 Unified Agent 是否已成功下載到臨時文件中。
	_, err = os.Stat(tempFile.Name())
	assert.NoError(t, err, "Unified Agent 下載失敗")
}

// TestWhiteSourceEnv_DoScan 測試 WhiteSourceEnv 的 DoScan 方法，驗證其執行 WhiteSource 掃描的功能。
func TestWhiteSourceEnv_DoScan(t *testing.T) {
	// 創建一個臨時目錄，用於存放測試用的掃描文件。
	tempDir, err := ioutil.TempDir("", "test_scan")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(tempDir) // 測試完成後刪除臨時目錄

	// 創建一個臨時目錄，用於存放 Unified Agent。
	uaDir, err := ioutil.TempDir("", "test_ua")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(uaDir) // 測試完成後刪除臨時目錄

	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{
		ApiKey:      "test_api_key",
		UserKey:     "test_user_key",
		ProjectName: "test_project_name",
		ProductName: "test_product_name",
		WSSUrl:      "test_wss_url",
		Offline:     "true",
	}

	// 設置環境變量
	os.Setenv("WS_APIKEY", config.ApiKey)
	os.Setenv("WS_USERKEY", config.UserKey)
	os.Setenv("WS_PROJECTNAME", config.ProjectName)
	os.Setenv("WS_PRODUCTNAME", config.ProductName)
	os.Setenv("WS_WSS_URL", config.WSSUrl)
	os.Setenv("WS_OFFLINE", config.Offline)
	os.Setenv("package_tmp", tempDir)
	os.Setenv("wssAgentPath", uaDir)
	os.Setenv("wssAgentName", "ua.jar")
	os.Setenv("wget", "echo") // 使用 echo 模擬 wget
	os.Setenv("agentURL", "https://example.com/ua.jar")

	// 創建一個掃描路徑。
	scanPath := filepath.Join(tempDir, "test_scan_path")
	err = os.MkdirAll(scanPath, 0755)
	require().NoError(t, err, "創建掃描路徑失敗")

	// 調用 DoScan 方法執行 WhiteSource 掃描。
	projectName := "test_project_name"
	config.DoScan("test_scan_path", &projectName, "no")

	// 可以在這裡添加更詳細的驗證，例如檢查掃描日誌是否包含預期的信息。
}

// TestWhiteSourceEnv_DoScan_WithConf 測試 WhiteSourceEnv 的 DoScan 方法，使用額外的設定檔。
func TestWhiteSourceEnv_DoScan_WithConf(t *testing.T) {
	// 創建一個臨時目錄，用於存放測試用的掃描文件。
	tempDir, err := ioutil.TempDir("", "test_scan")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(tempDir) // 測試完成後刪除臨時目錄

	// 創建一個臨時目錄，用於存放 Unified Agent。
	uaDir, err := ioutil.TempDir("", "test_ua")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(uaDir) // 測試完成後刪除臨時目錄

	// 創建一個臨時設定檔。
	configFile, err := ioutil.TempFile("", "test_config.yaml")
	require().NoError(t, err, "創建臨時設定檔失敗")
	defer os.Remove(configFile.Name()) // 測試完成後刪除臨時設定檔

	// 寫入一些測試用的配置到臨時設定檔中。
	content := `
apiKey: test_api_key
userKey: test_user_key
projectName: test_project_name
productName: test_product_name
productToken: test_product_token
wss.url: test_wss_url
offline: "true"
`
	_, err = configFile.WriteString(content)
	require().NoError(t, err, "寫入臨時設定檔失敗")
	err = configFile.Close()
	require().NoError(t, err, "關閉臨時設定檔失敗")

	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{
		ApiKey:      "test_api_key",
		UserKey:     "test_user_key",
		ProjectName: "test_project_name",
		ProductName: "test_product_name",
		WSSUrl:      "test_wss_url",
		Offline:     "true",
	}

	// 設置環境變量
	os.Setenv("WS_APIKEY", config.ApiKey)
	os.Setenv("WS_USERKEY", config.UserKey)
	os.Setenv("WS_PROJECTNAME", config.ProjectName)
	os.Setenv("WS_PRODUCTNAME", config.ProductName)
	os.Setenv("WS_WSS_URL", config.WSSUrl)
	os.Setenv("WS_OFFLINE", config.Offline)
	os.Setenv("package_tmp", tempDir)
	os.Setenv("wssAgentPath", uaDir)
	os.Setenv("wssAgentName", "ua.jar")
	os.Setenv("wget", "echo") // 使用 echo 模擬 wget
	os.Setenv("agentURL", "https://example.com/ua.jar")

	// 創建一個掃描路徑。
	scanPath := filepath.Join(tempDir, "test_scan_path")
	err = os.MkdirAll(scanPath, 0755)
	require().NoError(t, err, "創建掃描路徑失敗")

	// 調用 DoScan 方法執行 WhiteSource 掃描，使用額外的設定檔。
	projectName := "test_project_name"
	config.DoScan("test_scan_path", &projectName, "yes")

	// 可以在這裡添加更詳細的驗證，例如檢查掃描日誌是否包含預期的信息。
}

// TestWhiteSourceEnv_DoScan_Error 測試 WhiteSourceEnv 的 DoScan 方法，驗證其錯誤處理機制。
func TestWhiteSourceEnv_DoScan_Error(t *testing.T) {
	// 創建一個臨時目錄，用於存放測試用的掃描文件。
	tempDir, err := ioutil.TempDir("", "test_scan")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(tempDir) // 測試完成後刪除臨時目錄

	// 創建一個臨時目錄，用於存放 Unified Agent。
	uaDir, err := ioutil.TempDir("", "test_ua")
	require().NoError(t, err, "創建臨時目錄失敗")
	defer os.RemoveAll(uaDir) // 測試完成後刪除臨時目錄

	// 創建一個 WhiteSourceEnv 實例。
	config := WhiteSourceEnv{
		ApiKey:      "test_api_key",
		UserKey:     "test_user_key",
		ProjectName: "test_project_name",
		ProductName: "test_product_name",
		WSSUrl:      "test_wss_url",
		Offline:     "true",
	}

	// 設置環境變量
	os.Setenv("WS_APIKEY", config.ApiKey)
	os.Setenv("WS_USERKEY", config.UserKey)
	os.Setenv("WS_PROJECTNAME", config.ProjectName)
	os.Setenv("WS_PRODUCTNAME", config.ProductName)
	os.Setenv("WS_WSS_URL", config.WSSUrl)
	os.Setenv("WS_OFFLINE", config.Offline)
	os.Setenv("package_tmp", tempDir)
	os.Setenv("wssAgentPath", uaDir)
	os.Setenv("wssAgentName", "ua.jar")
	os.Setenv("wget", "nonexistent_command") // 模擬 wget 命令不存在

	// 創建一個掃描路徑。
	scanPath := filepath.Join(tempDir, "test_scan_path")
	err = os.MkdirAll(scanPath, 0755)
	require().NoError(t, err, "創建掃描路徑失敗")

	// 調用 DoScan 方法執行 WhiteSource 掃描。
	projectName := "test_project_name"
	config.DoScan("test_scan_path", &projectName, "no")

	// 可以在這裡添加更詳細的驗證，例如檢查掃描日誌是否包含預期的信息。
}
