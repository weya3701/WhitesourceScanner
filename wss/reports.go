package wss

import "encoding/json"

type InventoryReport struct {
	ProjectVitals ProjectVitals `json:"projectVitals" url:"projectVitals,omitempty"`
	Libraries     []Library     `json:"libraries" url:"libraries,omitempty"`
}

func (p *ProjectInfoRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	p.RequestType = "getProjectInventory"
	p.UserKey = uo.UserKey
	p.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
}

func (p ProjectInfoRequest) GetJsonData() ([]byte, error) {

	jsonData, err := json.Marshal(p)
	return jsonData, err
}

// func (pa *ProjectAlertRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
// 	projectName := ud.GetProjectName()
// 	pa.RequestType = "getProjectAlerts"
// 	pa.UserKey = uo.UserKey
// 	pa.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
// }
//
// func (pa ProjectAlertRequest) GetJsonData() ([]byte, error) {
// 	jsonData, err := json.Marshal(pa)
// 	return jsonData, err
// }

func (pir *ProjectInventoryRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	pir.RequestType = "getProjectInventory"
	pir.UserKey = uo.UserKey
	pir.Format = "xlsx"
	pir.ExtraLibraryFields = []string{"releaseDate"}

	pir.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
}

func (pir ProjectInventoryRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(pir)
	return jsonData, err
}

func (invent InventoryReport) GetReport() error {
	var err error = nil
	// Do function.

	// End function.
	return err
}

func (rr *ProjectRiskRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	rr.RequestType = "getProjectRiskReport"
	rr.UserKey = uo.UserKey
	rr.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
}

func (rr ProjectRiskRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(rr)
	return jsonData, err
}

func (p *AsyncProcessStatusRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	p.RequestType = "getAsyncProcessStatus"
	p.UserKey = uo.UserKey
}

func (p *AsyncProcessStatusRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(p)
	return jsonData, err
}

func (p *GenerateProjectReportAsyncRequest) InitRequest(uo UpdateRequestOriginal, ud UploadResponseData) {
	projectName := ud.GetProjectName()
	p.RequestType = "generateProjectReportAsync"
	p.UserKey = uo.UserKey
	p.ProjectToken = ud.ProjectNamesToDetails[projectName].ProjectToken
	p.ReportType = "ProjectInventoryReport"
}

func (p *GenerateProjectReportAsyncRequest) GetJsonData() ([]byte, error) {
	jsonData, err := json.Marshal(p)
	return jsonData, err
}
