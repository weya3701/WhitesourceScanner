package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"wss/inventory"
	"wss/worker"
	"wss/wss"

	"github.com/signintech/pdft"
	gopdf "github.com/signintech/pdft/minigopdf"
)

// initialPackageDefintion 根據 packageType 初始化並返回對應的 worker.Worker 實例。
func initialPackageDefintion(packageType string) worker.Worker {

	switch packageType {
	case "pip":
		return worker.Pypi{Command: os.Getenv("pip")}
	case "maven":
		return worker.Mvn{Command: os.Getenv("maven")}
	case "npm":
		return worker.Npm{Command: os.Getenv("npm")}
	case "gradle":
		return worker.Gradle{Command: os.Getenv("gradle")}
	case "wget":
		return worker.UrlGet{Command: os.Getenv("wget")}
	default:
		return worker.Pypi{Command: os.Getenv("pip")}
	}

}

// SyncDefintionPackages 同步定義檔中的套件。
// 它會根據 packageType 選擇合適的 worker 進行套件同步。
func SyncDefinitionPackages(packageType string, projectName string, requirementsFile string) error {
	wss.Verbosef("同步套件定義：type=%s project=%s file=%s", packageType, projectName, requirementsFile)
	var wk worker.WorkerHandler = worker.NewRepositoryWorker(initialPackageDefintion(packageType))
	if err := wk.SyncPackagesFromDefintionFile(projectName, requirementsFile); err != nil {
		return err
	}
	wss.Verbosef("套件定義同步完成：project=%s", projectName)
	return nil

}

// GetPackageReport 執行 WhiteSource 掃描，上傳請求，生成專案報告並獲取處理狀態，
// 最後取得專案風險報告。
func GetPackageReport(packageName string, projectName string, configFile string, directScanSource bool) (bool, error) {
	wss.Verbosef("[1/5] 開始掃描來源套件")
	if err := wss.DoWhitesourceScan(packageName, projectName, configFile, directScanSource); err != nil {
		return false, err
	}
	wss.Verbosef("[2/5] 上傳掃描結果")
	if _, err := wss.DoUploadRequest(projectName); err != nil {
		return false, err
	}

	wss.Verbosef("[3/5] 要求 Mend 產生專案報告")
	err, processID := wss.GenerateProjectReportAsync(projectName)
	if err != nil {
		return false, err
	}
	wss.Verbosef("[4/5] 等待專案報告完成：process_id=%s", processID)
	if _, err := wss.GetProcessStatus(processID, projectName); err != nil {
		return false, err
	}

	reportPath := fmt.Sprintf("report/%s", projectName)
	if err := os.MkdirAll(reportPath, 0755); err != nil {
		return false, fmt.Errorf("create report directory: %w", err)
	}
	wss.Verbosef("[5/5] 下載專案風險報告")
	if err := wss.GetProjectRiskReport(projectName); err != nil {
		return false, err
	}
	wss.Verbosef("套件掃描與報告取得完成：project=%s", projectName)
	return true, nil
}

func GetInventoryReport(projectName, packageType string) (bool, error) {
	urlSource := inventory.ReferenceURL
	switch packageType {
	case "maven", "gradle":
		urlSource = inventory.POMURL
	}

	source := fmt.Sprintf("%s/%s/alert.json", os.Getenv("report_tmp"), projectName)
	output := fmt.Sprintf("%s/%s/inventory.csv", os.Getenv("report_tmp"), projectName)
	wss.Verbosef("轉換庫存報告：source=%s output=%s url_source=%s", source, output, urlSource)
	if err := inventory.ConvertFile(source, output, urlSource); err != nil {
		return false, err
	}
	wss.Verbosef("庫存報告已儲存：%s", output)
	return true, nil
}

// GetProjectAlert 取得專案的風險警報，格式化並保存到 alert.json 檔案中。
func GetProjectAlert(projectName string) (bool, error) {
	var status bool = true
	var err error = nil
	wss.Verbosef("取得專案警報：project=%s", projectName)
	rsp, err := wss.GetProjectRiskAlert(projectName)
	if err != nil {
		return false, err
	}
	rsp, err = wss.GetPrettyString(rsp)
	if err != nil {
		return false, fmt.Errorf("format project alert: %w", err)
	}

	reportPath := fmt.Sprintf("report/%s", projectName)
	reportFile := fmt.Sprintf("%s", reportPath+"/alert.json")
	if err := os.MkdirAll(reportPath, 0755); err != nil {
		return false, fmt.Errorf("create report directory: %w", err)
	}
	err = os.WriteFile(reportFile, []byte(rsp), 0644)
	if err != nil {
		status = false
		return status, err
	}
	wss.Verbosef("專案警報已儲存：path=%s bytes=%d", reportFile, len(rsp))

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

	rsp, err := wss.GetProjectRiskAlert(projectName)
	if err != nil {
		return err
	}
	rsp, err = wss.GetPrettyString(rsp)
	if err != nil {
		return err
	}

	var projectScanInfo wss.ProjectScanInfo
	if err := json.Unmarshal([]byte(rsp), &projectScanInfo); err != nil {
		return err
	}

	// FIXME. 變更時間為utf+8 -- Start
	layout := "2006-01-02 15:04:05"
	secondsInHour := 60 * 60
	loc := time.FixedZone("CST", 8*secondsInHour)

	t, err := time.ParseInLocation(layout, projectScanInfo.ProjectVitals.LastUpdatedDate, time.UTC)
	if err != nil {
		return err
	}

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
	err = ipdf.Open(reportFile)
	if err != nil {
		return fmt.Errorf("PDF not found %w", err)
	}

	ipdf.AddFont("arial", "./ttf/angsa.ttf")
	ipdf.SetFont("arial", "", 20)
	ipdf.Insert(timestamp, 1, 302, -5, 100, 100, gopdf.Center|gopdf.Bottom)
	ipdf.Save(reportFile)

	return nil
}
