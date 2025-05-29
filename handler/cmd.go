package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
	"wss/wss"

	"github.com/signintech/pdft"
	gopdf "github.com/signintech/pdft/minigopdf"
)

func GetPackageReport(packageName string, projectName string, withConf string) {
	wss.DoWhitesourceScan(packageName, projectName, withConf)
	wss.DoUploadRequest(projectName)

	ch := wss.GenerateProjectReportAsync(projectName)
	_ = wss.GetProcessStatus(ch, projectName)

	reportPath := fmt.Sprintf("report/%s", projectName)
	os.Mkdir(reportPath, 0755)
	rsp := wss.GetProjectRiskReport(projectName)
	_, err := json.Marshal(rsp)
	if err != nil {
		panic(err)
	}
}

func GetInventoryReport(projectName string) {
	source := fmt.Sprintf("report/%s/alert.json", projectName)
	output := fmt.Sprintf("report/%s/inventory.csv", projectName)
	shellCommand := fmt.Sprintf("./utils/inventory2csv.sh %s %s", source, output)
	cmd := exec.Command("bash", "-c", shellCommand)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("命令執行失敗，狀態碼: %d", exitErr.ExitCode())
		} else {
			log.Fatalf("命令執行時發生錯誤: %v", err)
		}
	}
}

func GetProjectAlert(projectName string) {
	rsp := wss.GetProjectRiskAlert(projectName)
	rsp, _ = wss.GetPrettyString(rsp)

	var projectScanInfo wss.ProjectScanInfo

	_ = json.Unmarshal([]byte(rsp), &projectScanInfo)

	reportPath := fmt.Sprintf("report/%s", projectName)
	reportFile := fmt.Sprintf(reportPath + "/alert.json")
	os.Mkdir(reportPath, 0755)
	err := os.WriteFile(reportFile, []byte(rsp), 0644)
	if err != nil {
		panic(err)
	}

	// fmt.Print(rsp)
}

func UpdateRiskReport(projectName string) {

	var ipdf pdft.PDFt

	rsp := wss.GetProjectRiskAlert(projectName)
	rsp, _ = wss.GetPrettyString(rsp)

	var projectScanInfo wss.ProjectScanInfo
	_ = json.Unmarshal([]byte(rsp), &projectScanInfo)

	timestamp := "lastUpload:" + projectScanInfo.ProjectVitals.LastUpdatedDate + " GenReport:" + time.Now().Format("2006-01-02 15:04:05")
	fmt.Println("timestamp: ", timestamp)

	reportFile := fmt.Sprintf(
		"report/%s/risk.pdf",
		projectName,
	)
	fmt.Println("reportFile: ", reportFile)
	err := ipdf.Open(reportFile)
	if err != nil {
		fmt.Println("PDF not found")
	}

	ipdf.AddFont("arial", "./ttf/angsa.ttf")
	ipdf.SetFont("arial", "", 20)
	ipdf.Insert(timestamp, 1, 300, 10, 100, 100, gopdf.Center|gopdf.Bottom)
	ipdf.Save(reportFile)
}
