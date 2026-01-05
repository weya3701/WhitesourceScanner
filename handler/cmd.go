package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
	"wss/worker"
	"wss/wss"

	"github.com/signintech/pdft"
	gopdf "github.com/signintech/pdft/minigopdf"
)

func initialPackageDefintion(packageType string) worker.Worker {

	switch packageType {
	case "python":

		return worker.Pypi{}
	case "maven":
		return worker.Mvn{}
	case "npm":
		return worker.Npm{}
	case "gradle":
		return worker.Gradle{}
	default:
		return worker.Pypi{}
	}

}

func SyncDefintionPackages(packageType string, projectName string, requirementsFile string) {
	var wk worker.WorkerHandler = worker.NewRepositoryWorker(initialPackageDefintion(packageType))
	wk.SyncPackagesFromDefintionFile(projectName, requirementsFile)

}

func GetPackageReport(packageName string, projectName string, withConf string) {
	wss.DoWhitesourceScan(packageName, projectName, withConf)
	wss.DoUploadRequest(projectName)

	ch := wss.GenerateProjectReportAsync(projectName)
	_ = wss.GetProcessStatus(ch, projectName)

	report_path := os.Getenv("report_path")
	reportPath := fmt.Sprintf("%s/%s", report_path, projectName)
	if err := os.MkdirAll(reportPath, 0755); err != nil {
		fmt.Println("Make dir failed.", err)
	}
	rsp := wss.GetProjectRiskReport(projectName)
	_, err := json.Marshal(rsp)
	if err != nil {
		panic(err)
	}
}

func GetInventoryReport(projectName string) {
	report_path := os.Getenv("report_path")
	source := fmt.Sprintf("%s/%s/alert.json", report_path, projectName)
	output := fmt.Sprintf("%s/%s/inventory.csv", report_path, projectName)
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

	report_path := os.Getenv("report_path")
	reportPath := fmt.Sprintf("%s/%s", report_path, projectName)
	fmt.Println("reportPath: ", reportPath)
	if err := os.Mkdir(reportPath, 0755); err != nil {
		fmt.Println("Make dir failed.")
	} else {
		fmt.Println(reportPath)
	}

	reportFile := fmt.Sprintf(reportPath + "/alert.json")
	// os.Mkdir(reportPath, 0755)
	err := os.WriteFile(reportFile, []byte(rsp), 0644)
	fmt.Println(err)
	if err != nil {
		panic(err)
	}
}

// FIXME.
func InitMendCli(exportFile, application, packageName, projectName, tarFile, imageName, imageTag string) wss.MendCli {
	var mendCli wss.MendCli
	mendCli.Application = application
	mendCli.ExportFile = exportFile
	mendCli.ImageName = imageName
	mendCli.ImageTag = imageTag
	mendCli.TarFile = tarFile
	mendCli.PackageName = packageName
	mendCli.ProjectName = projectName

	return mendCli
}

func UpdateRiskReport(projectName string) {

	var ipdf pdft.PDFt

	rsp := wss.GetProjectRiskAlert(projectName)
	rsp, _ = wss.GetPrettyString(rsp)

	var projectScanInfo wss.ProjectScanInfo
	_ = json.Unmarshal([]byte(rsp), &projectScanInfo)

	timestamp := "lastUpload:" + projectScanInfo.ProjectVitals.LastUpdatedDate + " GenReport:" + time.Now().Format("2006-01-02 15:04:05")
	fmt.Println("timestamp: ", timestamp)
	report_path := os.Getenv("report_path")
	reportFile := fmt.Sprintf("%s/%s/risk.pdf", report_path, projectName)
	fmt.Println("reportFile: ", reportFile)
	err := ipdf.Open(reportFile)
	if err != nil {
		fmt.Println("PDF not found")
	}

	ipdf.AddFont("arial", "./ttf/angsa.ttf")
	ipdf.SetFont("arial", "", 20)
	ipdf.Insert(timestamp, 1, 302, -5, 100, 100, gopdf.Center|gopdf.Bottom)
	ipdf.Save(reportFile)
}
