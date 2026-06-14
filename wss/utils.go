package wss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

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
func DoWhitesourceScan(packagePath string, productName string, withConf string) {
	var wssEnv WhiteSourceEnv
	projectName := &productName

	wssEnv.ParserEnv(os.Getenv("settings_file"))
	wssEnv.SetProductName(&productName)
	wssEnv.SetProjectName(projectName)

	wssEnv.SetEnv()
	wssEnv.DoScan(packagePath, &productName, withConf)

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
	return fmt.Sprintf(
		"%s%s/%s",
		path,
		projectName,
		fileName,
	)
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
	updateRequestorigin := NewUpdateRequestFromFile(requestFile)

	resp, _ := updateRequestorigin.SendUploadRequest(
		os.Getenv("whitesource_agent"),
	)
	body, err := io.ReadAll(resp.Body)
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
	uploadResponseStatus.ToFile(responseStatusFile)
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
	uploadResponseData.ToFile(responseDataFile)

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
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var asyncProcessStatusRequest GenerateProjectReportAsyncRequest
	var processStatusResponse ProcessStatusResponse

	requestFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("request_file"),
	)
	responseStatusFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("response_status_file"),
	)
	updateRequestOrigin.FromFile(requestFile)
	uploadResponseStatus.FromFile(responseStatusFile)
	err := json.Unmarshal(
		[]byte(uploadResponseStatus.Data),
		&uploadResponseData,
	)
	if err != nil {
		return fmt.Errorf("Failed to json unmarshal"), "''"
	}

	asyncProcessStatusRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	asyncProcessStatusRequest.Format = "json"

	jsonData, _ := asyncProcessStatusRequest.GetJsonData()

	_, body := AskProcessStatus(jsonData)
	err = json.Unmarshal(body, &processStatusResponse)
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

	var rsp []byte = nil
	req, _ := http.NewRequest(
		"POST",
		os.Getenv("whitesource_api"),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set(GetJsonContentType())
	req.Header.Set("Charset", "utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request"), rsp
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return err, body
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
func GetProcessStatus(uuid string, projectName string) string {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var asyncProcessStatusRequest AsyncProcessStatusRequest
	var asyncProcessResponse ProcessStatusResponse

	requestFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("request_file"),
	)
	responseStatusFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("response_status_file"),
	)
	updateRequestOrigin.FromFile(requestFile)
	uploadResponseStatus.FromFile(responseStatusFile)
	err := json.Unmarshal(
		[]byte(uploadResponseStatus.Data),
		&uploadResponseData,
	)
	if err != nil {
		fmt.Println("Failed to json unmarshal")
	}
	asyncProcessStatusRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	for {
		asyncProcessStatusRequest.Uuid = uuid
		asyncProcessStatusRequest.OrgToken = os.Getenv("WS_APIKEY")
		jsonData, _ := asyncProcessStatusRequest.GetJsonData()
		_, body := AskProcessStatus(jsonData)
		json.Unmarshal(body, &asyncProcessResponse)

		if asyncProcessResponse.AsyncProcessStatus.Status == "SUCCESS" {
			return "SUCCESS" // 直接返回，優雅退出
		}
		time.Sleep(5 * time.Second)
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
func GetProjectRiskAlert(destination string) string {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var projectAlertRequest ProjectInfoRequest

	requestFile := GetFilePath(
		os.Getenv("whitesource_path"),
		destination,
		os.Getenv("request_file"),
	)
	responseStatusFile := GetFilePath(
		os.Getenv("whitesource_path"),
		destination,
		os.Getenv("response_status_file"),
	)
	updateRequestOrigin.FromFile(requestFile)
	uploadResponseStatus.FromFile(responseStatusFile)
	err := json.Unmarshal(
		[]byte(uploadResponseStatus.Data),
		&uploadResponseData,
	)
	if err != nil {
		fmt.Println("Failed to json unmarshal")
	}

	projectAlertRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, _ := projectAlertRequest.GetJsonData()
	_, body := AskProcessStatus(jsonData)

	return string(body)
}

// GetProjectRiskReport 獲取並儲存專案的風險報告。
// 它會載入現有的請求和回應資料，構建專案風險請求，發送請求，然後將回應主體寫入檔案。
//
// 參數:
//   - destination: 專案名稱，用於構建檔案路徑和報告儲存路徑。
//
// 返回:
//   - map[string]string: 包含操作狀態和狀態碼的映射。
func GetProjectRiskReport(destination string) map[string]string {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var projectRiskRequest ProjectRiskRequest

	requestFile := GetFilePath(
		os.Getenv("whitesource_path"),
		destination,
		os.Getenv("request_file"),
	)
	responseStatusFile := GetFilePath(
		os.Getenv("whitesource_path"),
		destination,
		os.Getenv("response_status_file"),
	)
	updateRequestOrigin.FromFile(requestFile)
	uploadResponseStatus.FromFile(responseStatusFile)
	err := json.Unmarshal(
		[]byte(uploadResponseStatus.Data),
		&uploadResponseData,
	)
	if err != nil {
		fmt.Println("Failed to json unmarshal")
	}

	projectRiskRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, _ := projectRiskRequest.GetJsonData()

	req, _ := http.NewRequest(
		"POST",
		os.Getenv("whitesource_api"),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set(GetJsonContentType())
	req.Header.Set("Charset", "utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request")
	}
	defer resp.Body.Close()

	dPath := fmt.Sprintf(
		"%s/%s/%s",
		os.Getenv("report_tmp"),
		destination,
		os.Getenv("risk_report_file"),
	)
	body, _ := io.ReadAll(resp.Body)
	err = os.WriteFile(
		dPath,
		body,
		0644,
	)
	if err != nil {
		return map[string]string{
			"status": "failed",
			"code":   "500",
		}
	}

	return map[string]string{
		"status": "successful",
		"code":   "200",
	}
}

// GetInventoryReport 獲取並解析庫存報告。
// 它會載入現有的請求和回應資料，構建專案庫存請求，發送請求，然後將回應主體解析為 InventoryReport 結構。
//
// 返回:
//   - InventoryReport: 解析後的庫存報告結構。
func GetInventoryReport() InventoryReport {
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
		fmt.Println("Failed to json unmarshal")
	}

	projectInventoryRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, _ := projectInventoryRequest.GetJsonData()

	req, _ := http.NewRequest(
		"POST",
		os.Getenv("whitesource_api"),
		bytes.NewBuffer(jsonData),
	)
	req.Header.Set(GetJsonContentType())
	req.Header.Set("Charset", "utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &inventoryReport)
	if err != nil {
		fmt.Println("Failed to json unmarshal")
	}

	return inventoryReport
}
