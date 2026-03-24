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

var once sync.Once
var scan_mutex *sync.Mutex

// scan_mutex:  這個變數用於儲存用於同步掃描操作的 mutex。它使用單例模式，確保只初始化一次。
// once: 是一個 sync.Once 實例，用於保證程式碼塊只執行一次。  這是實現單例模式的關鍵。
// GetScanSingleton 是一個函數，用於取得掃描操作的 mutex 實例（單例模式）。
//
// 說明：
// 1. **單例模式 (Singleton Pattern):**  此函數實現了單例模式。  這意味著它確保只創建一個 `sync.Mutex` 實例，並且每次調用 `GetScanSingleton()` 時都返回相同的實例。
// 2. **sync.Once 的用途:**  `sync.Once` 是一個同步原語，用於確保程式碼塊只被執行一次，即使在多個 goroutine 同時調用的情況下。
// 3. **如何工作:**
//   - `once.Do()` 接收一個函數作為參數。
//   - 第一次調用 `once.Do()` 時，它會執行傳遞的函數（在這個例子中是初始化 `scan_mutex`）。
//   - 之後的任何次調用 `once.Do()` 時，都會立即返回，而不會再次執行函數。
//
// 4. **保護共享資源:**  `sync.Mutex` 被用來保護對共享資源的並發訪問。 在掃描操作中，`scan_mutex` 可能用於保護對掃描結果、掃描配置或其他共享數據結構的訪問，以避免競爭條件和數據損壞。
//
// 返回值：
//   - `*sync.Mutex`:  返回指向掃描操作的 mutex 的指標。
func GetScanSingleton() *sync.Mutex {
	once.Do(
		func() {
			scan_mutex = &sync.Mutex{}
		},
	)
	return scan_mutex
}

func (config *WhiteSourceEnv) ParserEnv(fpath string) {

	data, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
}

func (w *WhiteSourceEnv) SetProjectName(projectName *string) {
	w.ProjectName = *projectName
}

func (w *WhiteSourceEnv) SetProductName(productName *string) {
	w.ProjectName = *productName
}

func (w WhiteSourceEnv) SetEnv() {
	os.Setenv("WS_APIKEY", w.ApiKey)
	os.Setenv("WS_USERKEY", w.UserKey)
	os.Setenv("WS_PROJECTNAME", w.ProjectName)
	os.Setenv("WS_PRODUCTNAME", w.ProductName)
	os.Setenv("WS_PRODUCTTOKEN", w.ProductToken)
	os.Setenv("WS_WSS_URL", w.WSSUrl)
	os.Setenv("WS_OFFLINE", w.Offline)
}

func MoveRequestFile(source string, destination string) {
	err := os.Rename(source, destination)
	if err != nil {
		panic(err)
	}
}

func CreateDirectory(base string, dirname string) {
	dir := fmt.Sprintf("%s/%s", base, dirname)
	os.Mkdir(dir, 0755)
}

type MendCli struct {
	ExportFile  string
	Application string
	PackageName string
	ProjectName string
	TarFile     string
	ImageName   string
	ImageTag    string
}

// If file is exists.
// 引件兩參數, filepath, filename，系統判斷目錄資料夾及檔案是否存在，若不存在則建立，若目錄已存在則開始執行指定命令，若有錯誤回傳錯誤
func initialUnifiedAgent(fpath, filename string) error {
	var err error = nil

	// 建立完整檔案路徑
	agentfilePath := filepath.Join(fpath, filename)

	// 檢查目錄是否存在
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		// 目錄不存在，建立目錄 (包含所有父目錄)
		if err := os.MkdirAll(fpath, 0755); err != nil { // 0755: 讀寫執行權限，可根據需要調整
			return fmt.Errorf("無法建立目錄 %s: %w", fpath, err)
		}
		fmt.Printf("目錄 %s 已建立\n", fpath)
	} else if err != nil {
		// 檢查目錄存在時，發生其他錯誤
		return fmt.Errorf("檢查目錄 %s 時發生錯誤: %w", fpath, err)
	} else {
		// 目錄已存在
		fmt.Printf("目錄 %s 已存在\n", fpath)

		// 在這裡執行你的指定命令。  這裡用印出訊息代替。
		// 例如:  執行一個檔案讀取、寫入、或任何其他操作。
		fmt.Printf("執行指定命令... (檔案為 %s)\n", agentfilePath) //  代表指定命令，可以替換為其他命令
	}

	// 檢查檔案是否存在 (可選，根據需求加入)
	fmt.Println("agent file: ", agentfilePath)
	if _, err := os.Stat(agentfilePath); os.IsNotExist(err) {
		// 檔案不存在，可以選擇在此建立檔案(如果需要)
		// file, err := os.Create(filePath)
		// if err != nil {
		// 	return fmt.Errorf("無法建立檔案 %s: %w", filePath, err)
		// }
		// defer file.Close()
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

// DoDockerTarFileScan 函數使用 Mend CLI 掃描 Docker 壓縮包文件。
// 它接受一個 MendCli 結構體作為參數，該結構體包含掃描所需的配置信息。
//
// 參數：
//
//	cli: MendCli 結構體，包含應用程序名稱、項目名稱和 Docker 壓縮包文件的路徑。
//
// 流程：
//  1. 根據 cli.Application 和 cli.ProjectName 構造掃描範圍字符串 (scanScope)。
//  2. 確定 Docker 壓縮包文件的路徑 (dockerTarFile)。
//     如果 cli.TarFile 為空，則默認使用 "/tmp/<項目名稱>/<項目名稱>.tar"。
//     否則，使用 cli.TarFile 指定的路徑。
//  3. 构造執行 Mend CLI 命令的參數 (cmdArgs)。
//  4. 使用 exec.Command 創建一個命令，以執行 Mend CLI，並將其標準輸出重定向到 stdout 緩衝區。
//  5. 執行命令。
//  6. 打印命令的標準輸出到控制台。
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
//
// 參數：
//   - packagePath: 掃描的套件路徑。
//   - projectName: 專案名稱的指標。  如果為 nil，則使用 w.ProjectName。
//   - withConf: 是否使用額外的設定檔。如果為 "yes"，則使用 wss-unified-agent.config。
//
// 流程：
//  1. 獲取掃描的單例鎖。
//  2. 建立掃描路徑 (./tmp/<packagePath>)。
//  3. 建立 WhiteSource 掃描命令的參數列表。
//  4. 如果 withConf 為 "yes"，則加入設定檔參數。
//  5. 建立一個帶有 10 分鐘超時的 context。
//  6. 建立執行 WhiteSource 掃描的命令 (使用 Java 和 wss-unified-agent.jar)。
//  7. 執行命令並獲取輸出和錯誤。
//  8. 打印命令的輸出。
//  9. 如果掃描失敗，打印錯誤訊息。
//  10. 建立 whitesource 目錄，使用 w.ProjectName 作為子目錄。
//  11. 建立更新請求檔案的目的地路徑 (whitesource/<w.ProjectName>/update-request.txt)。
//  12. 將更新請求檔案移動到目的地。
func (w WhiteSourceEnv) DoScan(packagePath string, projectName *string, withConf string) {

	// initial unified agent
	initialUnifiedAgent(os.Getenv("wssAgentPath"), os.Getenv("wssAgentName"))

	mutex := GetScanSingleton()
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
		fmt.Println("Scan failed.")
	}

	CreateDirectory("whitesource", w.ProjectName)
	destinationFile := fmt.Sprintf("whitesource/%s/update-request.txt", w.ProjectName)
	MoveRequestFile("whitesource/update-request.txt", destinationFile)
}
