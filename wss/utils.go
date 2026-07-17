package wss

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	httpTimeout     = 30 * time.Second
	pollTimeout     = 10 * time.Minute
	pollInterval    = 5 * time.Second
	maxResponseSize = 20 << 20
)

var apiHTTPClient = &http.Client{Timeout: httpTimeout}

// GetJsonContentType 返回 JSON 內容類型及其值。
//
// 返回:
//   - string: "Content-Type" 標頭名稱。
//   - string: "application/json" 內容類型值。
func GetJsonContentType() (string, string) {
	return "Content-Type", "application/json"
}

// DoWhitesourceScan 執行 WhiteSource 掃描流程。
// 它會解析 WhiteSource 環境配置，設定專案名稱和產品名稱，然後執行掃描。
//
// 參數:
//   - packagePath: 要掃描的套件路徑。
//   - productName: 產品名稱。
//   - withConf: 是否使用配置檔案 ("yes" 表示使用)。
func DoWhitesourceScan(packagePath string, productName string, withConf string) error {
	var wssEnv WhiteSourceEnv
	projectName := &productName

	if err := wssEnv.ParserEnv(os.Getenv("settings_file")); err != nil {
		return err
	}
	wssEnv.SetProductName(&productName)
	wssEnv.SetProjectName(projectName)

	if err := wssEnv.SetEnv(); err != nil {
		return err
	}
	return wssEnv.DoScan(packagePath, &productName, withConf)
}

// GetFilePath 根據提供的路徑、專案名稱和檔案名稱構建完整的檔案路徑。
//
// 參數:
//   - path: 基礎路徑。
//   - projectName: 專案名稱。
//   - fileName: 檔案名稱。
//
// 返回:
//   - string: 完整的檔案路徑。
func GetFilePath(path string, projectName string, fileName string) string {
	return filepath.Join(path, projectName, fileName)
}

func loadRequestData(projectName string) (UpdateRequestOriginal, UploadResponseData, error) {
	var request UpdateRequestOriginal
	var status UploadResponseStatus
	var data UploadResponseData
	requestFile := GetFilePath(os.Getenv("whitesource_path"), projectName, os.Getenv("request_file"))
	statusFile := GetFilePath(os.Getenv("whitesource_path"), projectName, os.Getenv("response_status_file"))
	if !request.FromFile(requestFile) {
		return request, data, fmt.Errorf("load update request file %s", requestFile)
	}
	if !status.FromFile(statusFile) {
		return request, data, fmt.Errorf("load upload response file %s", statusFile)
	}
	if err := json.Unmarshal([]byte(status.Data), &data); err != nil {
		return request, data, fmt.Errorf("decode upload response data: %w", err)
	}
	return request, data, nil
}

// DoUploadRequest 執行 WhiteSource 的上傳請求流程。
// 它會從檔案讀取原始更新請求，發送上傳請求，然後將回應狀態和回應資料寫入檔案。
//
// 參數:
//   - projectName: 專案名稱，用於構建檔案路徑。
//
// 返回:
//   - string: 成功時的訊息，或錯誤訊息。
//   - error: 如果發生任何錯誤，返回錯誤；否則返回 nil。
func DoUploadRequest(projectName string) (string, error) {
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var err error = nil
	var msg string = ""

	requestFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("request_file"),
	)
	updateRequestorigin, err := NewUpdateRequestFromFile(requestFile)
	if err != nil {
		return "Failed to load upload request", err
	}

	resp, err := updateRequestorigin.SendUploadRequest(
		os.Getenv("whitesource_agent"),
	)
	if err != nil {
		return "Failed to send upload request", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "Upload request failed", fmt.Errorf("upload API returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return "Failed to parse response body", err
	}
	err = json.Unmarshal(body, &uploadResponseStatus)
	if err != nil {
		return "Failed to parse response body", err
	}
	responseStatusFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("response_status_file"),
	)
	if !uploadResponseStatus.ToFile(responseStatusFile) {
		return "Failed to write response status", fmt.Errorf("write response status file %s", responseStatusFile)
	}
	datas := []byte(uploadResponseStatus.Data)
	err = json.Unmarshal(datas, &uploadResponseData)
	if err != nil {
		return "Failed to json unmarshal", err
	}
	responseDataFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("response_data_file"),
	)
	if !uploadResponseData.ToFile(responseDataFile) {
		return "Failed to write response data", fmt.Errorf("write response data file %s", responseDataFile)
	}

	return msg, err
}

// GenerateProjectReportAsync 啟動異步專案報告生成。
// 它會載入現有的請求和回應資料，初始化異步報告請求，然後發送請求並返回異步處理的 UUID。
//
// 參數:
//   - projectName: 專案名稱。
//
// 返回:
//   - error: 如果發生錯誤，返回錯誤。
//   - string: 異步處理的 UUID。
func GenerateProjectReportAsync(projectName string) (error, string) {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseData UploadResponseData
	var asyncProcessStatusRequest GenerateProjectReportAsyncRequest
	var processStatusResponse ProcessStatusResponse

	updateRequestOrigin, uploadResponseData, err := loadRequestData(projectName)
	if err != nil {
		return err, ""
	}

	asyncProcessStatusRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	asyncProcessStatusRequest.Format = "json"

	jsonData, err := asyncProcessStatusRequest.GetJsonData()
	if err != nil {
		return fmt.Errorf("encode report request: %w", err), ""
	}

	err, body := AskProcessStatus(jsonData)
	if err != nil {
		return err, ""
	}
	if err = json.Unmarshal(body, &processStatusResponse); err != nil {
		return fmt.Errorf("decode report response: %w", err), ""
	}
	if processStatusResponse.AsyncProcessStatus.Uuid == "" {
		return fmt.Errorf("report response did not include an async process UUID"), ""
	}
	return nil, processStatusResponse.AsyncProcessStatus.Uuid
}

// AskProcessStatus 向 WhiteSource API 發送請求以查詢異步處理狀態。
//
// 參數:
//   - jsonData: 包含請求詳細資訊的 JSON 格式位元組陣列。
//
// 返回:
//   - error: 如果發送請求失敗，返回錯誤。
//   - []byte: API 回應的主體內容。
func AskProcessStatus(jsonData []byte) (error, []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()
	return askProcessStatus(ctx, jsonData)
}

func askProcessStatus(ctx context.Context, jsonData []byte) (error, []byte) {
	var rsp []byte = nil
	req, err := http.NewRequestWithContext(ctx,
		"POST",
		os.Getenv("whitesource_api"),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("create API request: %w", err), rsp
	}
	req.Header.Set(GetJsonContentType())
	req.Header.Set("Charset", "utf-8")

	resp, err := apiHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send API request: %w", err), rsp
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return fmt.Errorf("read API response: %w", err), rsp
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned %s: %s", resp.Status, strings.TrimSpace(string(body))), body
	}
	return nil, body
}

// GetProcessStatus 輪詢異步處理狀態，直到狀態變為 "SUCCESS"。
// 它會載入現有的請求和回應資料，構建異步狀態請求，然後重複查詢 API 直到成功。
//
// 參數:
//   - uuid: 異步處理的 UUID。
//   - projectName: 專案名稱。
//
// 返回:
//   - string: 處理完成後的回應狀態 (例如 "SUCCESS")。
func GetProcessStatus(uuid string, projectName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
	defer cancel()
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseData UploadResponseData
	var asyncProcessStatusRequest AsyncProcessStatusRequest
	var asyncProcessResponse ProcessStatusResponse

	updateRequestOrigin, uploadResponseData, err := loadRequestData(projectName)
	if err != nil {
		return "", fmt.Errorf("decode upload response data: %w", err)
	}
	asyncProcessStatusRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		asyncProcessStatusRequest.Uuid = uuid
		asyncProcessStatusRequest.OrgToken = os.Getenv("WS_APIKEY")
		jsonData, err := asyncProcessStatusRequest.GetJsonData()
		if err != nil {
			return "", fmt.Errorf("encode process status request: %w", err)
		}
		err, body := askProcessStatus(ctx, jsonData)
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(body, &asyncProcessResponse); err != nil {
			return "", fmt.Errorf("decode process status response: %w", err)
		}

		status := strings.ToUpper(asyncProcessResponse.AsyncProcessStatus.Status)
		switch status {
		case "SUCCESS":
			return status, nil
		case "FAILURE", "FAILED", "ERROR", "CANCELLED", "CANCELED":
			return status, fmt.Errorf("report generation ended with status %s", status)
		}
		select {
		case <-ctx.Done():
			return status, fmt.Errorf("waiting for report generation: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// GetPrettyString 將 JSON 字串格式化為帶有縮排的易讀形式。
//
// 參數:
//   - str: 要格式化的 JSON 字串。
//
// 返回:
//   - string: 格式化後的 JSON 字串。
//   - error: 如果 JSON 解析失敗，返回錯誤。
func GetPrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer

	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}

// GetProjectRiskAlert 獲取專案的風險警報。
// 它會載入現有的請求和回應資料，構建專案資訊請求，然後發送請求並返回警報內容。
//
// 參數:
//   - destination: 專案名稱，用於構建檔案路徑。
//
// 返回:
//   - string: 包含專案風險警報的 JSON 字串。
func GetProjectRiskAlert(destination string) (string, error) {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseData UploadResponseData
	var projectAlertRequest ProjectInfoRequest

	updateRequestOrigin, uploadResponseData, err := loadRequestData(destination)
	if err != nil {
		return "", fmt.Errorf("decode upload response data: %w", err)
	}

	projectAlertRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, err := projectAlertRequest.GetJsonData()
	if err != nil {
		return "", fmt.Errorf("encode project alert request: %w", err)
	}
	err, body := AskProcessStatus(jsonData)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetProjectRiskReport 獲取並儲存專案的風險報告。
// 它會載入現有的請求和回應資料，構建專案風險請求，發送請求，然後將回應主體寫入檔案。
//
// 參數:
//   - destination: 專案名稱，用於構建檔案路徑和報告儲存路徑。
//
// 返回:
//   - map[string]string: 包含操作狀態和狀態碼的映射。
func GetProjectRiskReport(destination string) error {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseData UploadResponseData
	var projectRiskRequest ProjectRiskRequest

	updateRequestOrigin, uploadResponseData, err := loadRequestData(destination)
	if err != nil {
		return fmt.Errorf("decode upload response data: %w", err)
	}

	projectRiskRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, err := projectRiskRequest.GetJsonData()
	if err != nil {
		return fmt.Errorf("encode project risk request: %w", err)
	}
	err, body := AskProcessStatus(jsonData)
	if err != nil {
		return err
	}

	dPath := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("report_tmp"),
		destination,
		os.Getenv("risk_report_file"),
	)
	if err := os.MkdirAll(filepath.Dir(dPath), 0755); err != nil {
		return fmt.Errorf("create risk report directory: %w", err)
	}
	err = os.WriteFile(
		dPath,
		body,
		0644,
	)
	if err != nil {
		return fmt.Errorf("write project risk report %s: %w", dPath, err)
	}
	return nil
}

// GetInventoryReport 獲取並解析庫存報告。
// 它會載入現有的請求和回應資料，構建專案庫存請求，發送請求，然後將回應主體解析為 InventoryReport 結構。
//
// 返回:
//   - InventoryReport: 解析後的庫存報告結構。
func GetInventoryReport() (InventoryReport, error) {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var projectInventoryRequest ProjectInventoryRequest
	var inventoryReport InventoryReport

	requestFile := fmt.Sprintf(
		"%s%s",
		os.Getenv("whitesource_path"),
		os.Getenv("request_file"),
	)
	responseStatusFile := fmt.Sprintf(
		"%s%s",
		os.Getenv("whitesource_path"),
		os.Getenv("response_status_file"),
	)
	updateRequestOrigin.FromFile(requestFile)
	uploadResponseStatus.FromFile(responseStatusFile)
	err := json.Unmarshal(
		[]byte(uploadResponseStatus.Data),
		&uploadResponseData,
	)
	if err != nil {
		return inventoryReport, fmt.Errorf("decode upload response data: %w", err)
	}

	projectInventoryRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, err := projectInventoryRequest.GetJsonData()
	if err != nil {
		return inventoryReport, fmt.Errorf("encode inventory request: %w", err)
	}
	err, body := AskProcessStatus(jsonData)
	if err != nil {
		return inventoryReport, err
	}

	err = json.Unmarshal(body, &inventoryReport)
	if err != nil {
		return inventoryReport, fmt.Errorf("decode inventory response: %w", err)
	}

	return inventoryReport, nil
}
