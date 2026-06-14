package wss

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// scanMutexMap 用於儲存每個 taskId 關聯的互斥鎖，以確保掃描操作的執行順序。
var scanMutexMap sync.Map

// GetScanSingleton 根據 taskId 獲取一個單例互斥鎖。
// 如果該 taskId 尚無互斥鎖，則會創建一個並儲存。
//
// 參數:
//   - taskId: 用於識別特定掃描任務的唯一 ID。
//
// 返回:
//   - *sync.Mutex: 與 taskId 關聯的互斥鎖。
func GetScanSingleton(taskId string) *sync.Mutex {
	mutex, _ := scanMutexMap.LoadOrStore(taskId, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

// ParserEnv 從 YAML 檔案路徑解析環境配置到 WhiteSourceEnv 結構中。
//
// 參數:
//   - fpath: YAML 配置檔案的路徑。
func (config *WhiteSourceEnv) ParserEnv(fpath string) {

	data, err := os.ReadFile(fpath)
	if err != nil {
		fmt.Printf("Read file failed: %s", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Yaml unmarshal failed: %s", err)
	}
}

// SetProjectName 設定 WhiteSourceEnv 實例的 ProjectName。
//
// 參數:
//   - projectName: 指向專案名稱字符串的指標。
func (w *WhiteSourceEnv) SetProjectName(projectName *string) {
	w.ProjectName = *projectName
}

// SetProductName 設定 WhiteSourceEnv 實例的 ProductName。
//
// 參數:
//   - productName: 指向產品名稱字符串的指標。
func (w *WhiteSourceEnv) SetProductName(productName *string) {
	w.ProductName = *productName
}

// SetEnv 將 WhiteSourceEnv 結構中的值設定為環境變數。
// 這些環境變數包括 WS_APIKEY, WS_USERKEY, WS_PROJECTNAME, WS_PRODUCTNAME, WS_PRODUCTTOKEN, WS_WSS_URL, WS_OFFLINE。
func (w WhiteSourceEnv) SetEnv() {
	os.Setenv("WS_APIKEY", w.ApiKey)
	os.Setenv("WS_USERKEY", w.UserKey)
	os.Setenv("WS_PROJECTNAME", w.ProjectName)
	os.Setenv("WS_PRODUCTNAME", w.ProductName)
	os.Setenv("WS_PRODUCTTOKEN", w.ProductToken)
	os.Setenv("WS_WSS_URL", w.WSSUrl)
	os.Setenv("WS_OFFLINE", w.Offline)
}

// MoveRequestFile 將源檔案重新命名 (移動) 到目標路徑。
//
// 參數:
//   - source: 源檔案路徑。
//   - destination: 目標檔案路徑。
//
// 返回:
//   - error: 如果重新命名操作失敗，返回錯誤；否則返回 nil。
func MoveRequestFile(source string, destination string) error {
	err := os.Rename(source, destination)
	if err != nil {
		return fmt.Errorf("Rename filed: %s", err)

	}
	return err
}

// CreateDirectory 在指定基礎路徑下建立一個新目錄。
// 即使目錄已存在，也不會返回錯誤。
//
// 參數:
//   - base: 基礎路徑。
//   - dirname: 要建立的目錄名稱。
func CreateDirectory(base string, dirname string) {
	dir := fmt.Sprintf("%s/%s", base, dirname)
	os.Mkdir(dir, 0755)
}

// MendCli 結構體包含用於 Mend CLI 掃描的各種參數。
type MendCli struct {
	ExportFile  string // 導出檔案的路徑。
	Application string // 應用程式名稱。
	PackageName string // 套件名稱。
	ProjectName string // 專案名稱。
	TarFile     string // Docker tar 檔案的路徑。
	ImageName   string // Docker 映像檔名稱。
	ImageTag    string // Docker 映像檔標籤。
}

// initialUnifiedAgent 初始化並下載統一代理 (Unified Agent) 檔案。
// 它會檢查目標目錄和代理檔案是否存在，如果不存在則會創建目錄並下載代理。
//
// 參數:
//   - fpath: 統一代理的存放目錄。
//   - filename: 統一代理檔案的名稱。
//
// 返回:
//   - error: 如果在目錄或檔案操作中發生錯誤，或下載失敗，返回錯誤；否則返回 nil。
func initialUnifiedAgent(fpath, filename string) error {
	var err error = nil

	agentfilePath := filepath.Join(fpath, filename)

	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		if err := os.MkdirAll(fpath, 0755); err != nil { // 0755: 讀寫執行權限，可根據需要調整
			return fmt.Errorf("無法建立目錄 %s: %w", fpath, err)
		}
		fmt.Printf("目錄 %s 已建立\n", fpath)
	} else if err != nil {
		return fmt.Errorf("檢查目錄 %s 時發生錯誤: %w", fpath, err)
	} else {
		fmt.Printf("目錄 %s 已存在\n", fpath)

		fmt.Printf("執行指定命令... (檔案為 %s)\n", agentfilePath) //  代表指定命令，可以替換為其他命令
	}

	fmt.Println("agent file: ", agentfilePath)
	if _, err := os.Stat(agentfilePath); os.IsNotExist(err) {
		fmt.Printf("檔案 %s 不存在，您可以建立它\n", agentfilePath)
		initAgentErr := getUnifiedAgent(
			filename,
			os.Getenv("wget"),
			[]string{
				os.Getenv("agentURL"),
				"-o",
				fmt.Sprintf("%s/%s", fpath, filename),
			},
		)
		if initAgentErr != nil {
			err = initAgentErr
		}
	} else if err != nil {
		return fmt.Errorf("檢查檔案 %s 時發生錯誤: %w", agentfilePath, err)
	} else {
		fmt.Printf("檔案 %s 已存在\n", agentfilePath)
	}
	return err
}

// getUnifiedAgent 執行命令下載統一代理 (Unified Agent) 檔案。
// 它會將命令的輸出結果寫入到指定的檔案。
//
// 參數:
//   - filename: 輸出檔案的名稱，下載的代理將寫入此檔案。
//   - prefix: 下載工具的命令前綴 (例如 "wget")。
//   - cmds: 要執行的命令參數，應包含下載 URL 和輸出路徑。
//
// 返回:
//   - error: 如果命令執行失敗或寫入檔案失敗，返回錯誤；否則返回 nil。
func getUnifiedAgent(filename string, prefix string, cmds []string) error {

	var err error = nil

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, prefix, cmds...)
	fmt.Println("get unifed agent: ", cmd)
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

// DoDockerTarFileScan 使用 Mend CLI 執行 Docker Tar 檔案掃描。
// 它會根據 MendCli 結構中的參數構建並執行掃描命令。
//
// 參數:
//   - cli: 包含掃描所需參數的 MendCli 結構實例。
func DoDockerTarFileScan(cli MendCli) {
	scanScope := fmt.Sprintf("\"%s//%s\"", cli.Application, cli.ProjectName)
	var dockerTarFile string
	if cli.TarFile == "" {
		dockerTarFile = fmt.Sprintf("./tmp/%s/%s.tar", cli.ProjectName, cli.ProjectName)
	} else {
		dockerTarFile = cli.TarFile
	}

	cmdArgs := []string{"mend", "image", "--tar", dockerTarFile, "-s", scanScope}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Run()
	fmt.Println(stdout.String())
}

// DoScan 執行 WhiteSource 掃描。
// 它首先會初始化統一代理，然後根據提供的套件路徑和配置執行掃描。
// 掃描完成後，會將生成的請求檔案移動到指定的專案目錄下。
//
// 參數:
//   - packagePath: 要掃描的套件路徑。
//   - projectName: 指向專案名稱字符串的指標，用於獲取掃描互斥鎖。
//   - withConf: 字符串，指示是否使用配置檔案 ("yes" 表示使用)。
//
// 返回:
//   - error: 如果掃描失敗或檔案操作失敗，返回錯誤；否則返回 nil。
func (w WhiteSourceEnv) DoScan(packagePath string, projectName *string, withConf string) error {
	var err error = nil

	// initial unified agent
	initialUnifiedAgent(os.Getenv("wssAgentPath"), os.Getenv("wssAgentName"))

	mutex := GetScanSingleton(*projectName)
	mutex.Lock()
	defer mutex.Unlock()
	scanPath := fmt.Sprintf("%s/%s", os.Getenv("package_tmp"), packagePath)

	ua := fmt.Sprintf("%s%s", os.Getenv("wssAgentPath"), os.Getenv("wssAgentName"))
	cmdArgs := []string{"java", "-jar", ua, "-d", scanPath}
	// cmdArgs := []string{"java", "-jar", "./wss-unified-agent.jar", "-d", scanPath}
	if withConf == "yes" {
		cmdArgs = append(cmdArgs, "-c", "./config/wss-unified-agent.config")
		fmt.Println("command args: ", cmdArgs)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return fmt.Errorf("Scan failed.")
	}

	CreateDirectory("whitesource", w.ProjectName)
	destinationFile := fmt.Sprintf("whitesource/%s/update-request.txt", w.ProjectName)
	MoveRequestFile("whitesource/update-request.txt", destinationFile)
	return err
}
