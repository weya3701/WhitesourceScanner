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

func DoDockerTarFileScan(cli MendCli) {
	// func DoMendCliScan(cli MendCli) {
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
	// cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	// cmd.Run()
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		fmt.Println("Scan failed.")
	}

	CreateDirectory("whitesource", w.ProjectName)
	destinationFile := fmt.Sprintf("whitesource/%s/update-request.txt", w.ProjectName)
	MoveRequestFile("whitesource/update-request.txt", destinationFile)
}
