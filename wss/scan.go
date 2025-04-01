package wss

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

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

func (w WhiteSourceEnv) DoScan(packagePath string, projectName *string, withConf string) {

	mutex := GetScanSingleton()
	mutex.Lock()
	defer mutex.Unlock()
	scanPath := fmt.Sprintf("./tmp/%s", packagePath)
	cmd := exec.Command(
		"java",
		"-jar",
		"./wss-unified-agent.jar",
		// "-c",
		// "./config/wss-unified-agent.config",
		"-d",
		scanPath,
	)
	if withConf == "yes" {
		cmd = exec.Command(
			"java",
			"-jar",
			"./wss-unified-agent.jar",
			"-c",
			"./config/wss-unified-agent.config",
			"-d",
			scanPath,
		)
	}
	cmd.Run()

	CreateDirectory("whitesource", w.ProjectName)
	destinationFile := fmt.Sprintf("whitesource/%s/update-request.txt", w.ProjectName)
	MoveRequestFile("whitesource/update-request.txt", destinationFile)
}
