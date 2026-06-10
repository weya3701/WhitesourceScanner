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

// initialPackageDefintion 根據 packageType 初始化並返回對應的 worker.Worker 實例。
func initialPackageDefintion(packageType string) worker.Worker {

	switch packageType {
	case "python":

		return worker.Pypi{}
	case "maven":
		return worker.Mvn{Command: os.Getenv("maven")}
	case "npm":
		return worker.Npm{Command: os.Getenv("npm")}
	case "gradle":
		return worker.Gradle{Command: os.Getenv("gradle")}
	case "wget":
		return worker.UrlGet{Command: os.Getenv("wget")}
	default:
		return worker.Pypi{Command: os.Getenv("pypi")}
	}

}

// SyncDefintionPackages 同步定義檔中的套件。
// 它會根據 packageType 選擇合適的 worker 進行套件同步。
func SyncDefinitionPackages(packageType string, projectName string, requirementsFile string) {
	var wk worker.WorkerHandler = worker.NewRepositoryWorker(initialPackageDefintion(packageType))
	wk.SyncPackagesFromDefintionFile(projectName, requirementsFile)

}

// GetPackageReport 執行 WhiteSource 掃描，上傳請求，生成專案報告並獲取處理狀態，
// 最後取得專案風險報告。
func GetPackageReport(packageName string, projectName string, withConf string) (bool, error) {
	var status bool = true
	var err error = nil

	wss.DoWhitesourceScan(packageName, projectName, withConf)
	wss.DoUploadRequest(projectName)

	_, ch := wss.GenerateProjectReportAsync(projectName)
	_ = wss.GetProcessStatus(ch, projectName)

	reportPath := fmt.Sprintf("report/%s", projectName)
	os.Mkdir(reportPath, 0755)
	rsp := wss.GetProjectRiskReport(projectName)
	_, err = json.Marshal(rsp)
	if err != nil {
		log.Printf("Failed to json marshal %s", err)
		status = false
		return status, err
	}
	return status, err
}

func GetInventoryReport(projectName, packageType string) (bool, error) {

	var status bool = true
	var err error = nil
	var shellScript string = ""
	switch packageType {
	case "python":
		shellScript = "inventory2csv.sh"
	case "maven":
		shellScript = "inventory2csv_pom.sh"
	case "npm":
		shellScript = "inventory2csv.sh"
	case "gradle":
		shellScript = "inventory2csv_pom.sh"
	case "wget":
		shellScript = "inventory2csv.sh"
	default:
		shellScript = "inventory2csv.sh"
	}

	source := fmt.Sprintf("%s/%s/alert.json", os.Getenv("report_tmp"), projectName)
	output := fmt.Sprintf("%s/%s/inventory.csv", os.Getenv("report_tmp"), projectName)

	// 檢查源文件是否存在
	if _, err := os.Stat(source); os.IsNotExist(err) {
		log.Printf("源文件不存在: %s", source)
		return false, nil // 檔案不存在，不應視為致命錯誤，但操作失敗
	}

	shellCommand := fmt.Sprintf("./utils/%s %s %s", shellScript, source, output)
	cmd := exec.Command("bash", "-c", shellCommand)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("命令執行失敗，狀態碼: %d, 輸出: %s, 錯誤: %s", exitErr.ExitCode(), stdoutBuf.String(), stderrBuf.String())
			status = false
		} else {
			log.Fatalf("命令執行時發生錯誤: %v, 輸出: %s, 錯誤: %s", err, stdoutBuf.String(), stderrBuf.String())
			status = false
		}
	}
	return status, err
}

// GetProjectAlert 取得專案的風險警報，格式化並保存到 alert.json 檔案中。
func GetProjectAlert(projectName string) (bool, error) {
	var status bool = true
	var err error = nil
	rsp := wss.GetProjectRiskAlert(projectName)
	rsp, _ = wss.GetPrettyString(rsp)

	var projectScanInfo wss.ProjectScanInfo

	_ = json.Unmarshal([]byte(rsp), &projectScanInfo)

	reportPath := fmt.Sprintf("report/%s", projectName)
	reportFile := fmt.Sprintf("%s", reportPath+"/alert.json")
	os.Mkdir(reportPath, 0755)
	err = os.WriteFile(reportFile, []byte(rsp), 0644)
	if err != nil {
		status = false
		return status, err
	}

	return status, err
}

// InitMendCli 初始化並返回一個 MendCli 結構，設定各種掃描和報告相關的參數。
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

// UpdateRiskReport 獲取專案風險警報，並更新指定的 PDF 報告檔案，
// 主要更新報告中的時間戳為 UTC+8。
func UpdateRiskReport(projectName string) error {

	var ipdf pdft.PDFt

	rsp := wss.GetProjectRiskAlert(projectName)
	rsp, _ = wss.GetPrettyString(rsp)

	var projectScanInfo wss.ProjectScanInfo
	_ = json.Unmarshal([]byte(rsp), &projectScanInfo)

	// FIXME. 變更時間為utf+8 -- Start
	layout := "2006-01-02 15:04:05"
	secondsInHour := 60 * 60
	loc := time.FixedZone("CST", 8*secondsInHour)

	t, _ := time.ParseInLocation(layout, projectScanInfo.ProjectVitals.LastUpdatedDate, time.UTC)

	tInUTC8 := t.In(loc)
	timeStr := tInUTC8.Format(layout)

	// FIXME. 變更時間為utf+8 -- End

	timestamp := "lastUpload:" + timeStr + " GenReport:" + time.Now().Format("2006-01-02 15:04:05")

	reportFile := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("report_tmp"),
		projectName,
		os.Getenv("risk_report_file"),
	)
	err := ipdf.Open(reportFile)
	if err != nil {
		return fmt.Errorf("PDF not found %w", err)
	}

	ipdf.AddFont("arial", "./ttf/angsa.ttf")
	ipdf.SetFont("arial", "", 20)
	ipdf.Insert(timestamp, 1, 302, -5, 100, 100, gopdf.Center|gopdf.Bottom)
	ipdf.Save(reportFile)

	return nil
}
