package wss

import "encoding/json"

// InventoryReport 結構體表示庫存報告的詳細資訊。
type InventoryReport struct {
	ProjectVitals ProjectVitals `json:"projectVitals" url:"projectVitals,omitempty"` // 專案重要資訊
	Libraries     []Library     `json:"libraries" url:"libraries,omitempty"`         // 庫存中的函式庫列表
}

// InitRequest 初始化 ProjectInfoRequest 請求。
// 根據提供的原始更新請求和上傳回應資料，設定請求類型、用戶金鑰和專案令牌。
func (p *ProjectInfoRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	p.RequestType = "getProjectInventory"
	p.UserKey = uo.UserKey
	p.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
}

// GetJsonData 將 ProjectInfoRequest 實例序列化為 JSON 格式的位元組陣列。
func (p ProjectInfoRequest) GetJsonData() ([]byte, error) {

	jsonData, err := json.Marshal(p)
	return jsonData, err
}

// InitRequest 初始化 ProjectInventoryRequest 請求。
// 根據提供的原始更新請求和上傳回應資料，設定請求類型、用戶金鑰、報告格式、額外函式庫欄位和專案令牌。
func (pir *ProjectInventoryRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	pir.RequestType = "getProjectInventory"
	pir.UserKey = uo.UserKey
	pir.Format = "xlsx"
	pir.ExtraLibraryFields = []string{"releaseDate"}

	pir.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
}

// GetJsonData 將 ProjectInventoryRequest 實例序列化為 JSON 格式的位元組陣列。
func (pir ProjectInventoryRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(pir)
	return jsonData, err
}

// GetReport 是一個佔位符函式，用於獲取庫存報告。
// 目前未實作具體功能，僅返回 nil 錯誤。
func (invent InventoryReport) GetReport() error {
	var err error = nil
	// Do function.

	// End function.
	return err
}

// InitRequest 初始化 ProjectRiskRequest 請求。
// 根據提供的原始更新請求和上傳回應資料，設定請求類型、用戶金鑰和專案令牌。
func (rr *ProjectRiskRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	rr.RequestType = "getProjectRiskReport"
	rr.UserKey = uo.UserKey
	rr.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
}

// GetJsonData 將 ProjectRiskRequest 實例序列化為 JSON 格式的位元組陣列。
func (rr ProjectRiskRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(rr)
	return jsonData, err
}

// InitRequest 初始化 AsyncProcessStatusRequest 請求。
// 根據提供的原始更新請求和上傳回應資料，設定請求類型和用戶金鑰。
func (p *AsyncProcessStatusRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	p.RequestType = "getAsyncProcessStatus"
	p.UserKey = uo.UserKey
}

// GetJsonData 將 AsyncProcessStatusRequest 實例序列化為 JSON 格式的位元組陣列。
func (p *AsyncProcessStatusRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(p)
	return jsonData, err
}

// InitRequest 初始化 GenerateProjectReportAsyncRequest 請求。
// 根據提供的原始更新請求和上傳回應資料，設定請求類型、用戶金鑰、專案令牌和報告類型。
func (p *GenerateProjectReportAsyncRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	p.RequestType = "generateProjectReportAsync"
	p.UserKey = uo.UserKey
	p.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
	p.ReportType = "ProjectInventoryReport"
}

// GetJsonData 將 GenerateProjectReportAsyncRequest 實例序列化為 JSON 格式的位元組陣列。
func (p *GenerateProjectReportAsyncRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(p)
	return jsonData, err
}
