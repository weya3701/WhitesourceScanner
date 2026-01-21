package wss

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
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

func (config *WhiteSourceEnv) ParserEnv(filepath string) {

	data, err := os.ReadFile(filepath)
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

	mutex := GetScanSingleton()
	mutex.Lock()
	defer mutex.Unlock()
	scanPath := fmt.Sprintf("./tmp/%s", packagePath)

	cmdArgs := []string{"java", "-jar", "./wss-unified-agent.jar", "-d", scanPath}
	if withConf == "yes" {
		cmdArgs = append(cmdArgs, "-c", "./config/wss-unified-agent.config")
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
