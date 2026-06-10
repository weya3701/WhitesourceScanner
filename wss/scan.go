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

var scanMutexMap sync.Map

func GetScanSingleton(taskId string) *sync.Mutex {
	mutex, _ := scanMutexMap.LoadOrStore(taskId, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

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

func MoveRequestFile(source string, destination string) error {
	err := os.Rename(source, destination)
	if err != nil {
		return fmt.Errorf("Rename filed: %s", err)

	}
	return err
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
