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

func GetJsonContentType() (string, string) {
	return "Content-Type", "application/json"
}

func DoWhitesourceScan(packagePath string, productName string, withConf string) {
	var wssEnv WhiteSourceEnv

	wssEnv.ParserEnv(os.Getenv("settings_file"))
	wssEnv.SetProductName(&productName)

	wssEnv.SetEnv()
	wssEnv.DoScan(packagePath, &productName, withConf)

}

func GetFilePath(path string, projectName string, fileName string) string {
	return fmt.Sprintf(
		"%s%s/%s",
		path,
		projectName,
		fileName,
	)
}

func DoUploadRequest(projectName string) {
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData

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
	print(string(body))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &uploadResponseStatus)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	responseDataFile := GetFilePath(
		os.Getenv("whitesource_path"),
		projectName,
		os.Getenv("response_data_file"),
	)
	uploadResponseData.ToFile(responseDataFile)

}

func GenerateProjectReportAsync(projectName string) string {
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
		panic(err)
	}

	asyncProcessStatusRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	asyncProcessStatusRequest.Format = "json"

	jsonData, _ := asyncProcessStatusRequest.GetJsonData()

	body := AskProcessStatus(jsonData)
	err = json.Unmarshal(body, &processStatusResponse)
	return processStatusResponse.AsyncProcessStatus.Uuid
}

func AskProcessStatus(jsonData []byte) []byte {
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
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}

func GetProcessStatus(uuid string, projectName string) string {
	var updateRequestOrigin UpdateRequestOriginal
	var uploadResponseStatus UploadResponseStatus
	var uploadResponseData UploadResponseData
	var asyncProcessStatusRequest AsyncProcessStatusRequest
	var asyncProcessResponse ProcessStatusResponse
	var status string = "no"

	ch := make(chan string)
	stop := make(chan string)

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
		panic(err)
	}
	asyncProcessStatusRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	for {
		if status == "no" {
			go func(uuid string) {
				ch <- uuid
			}(uuid)
		} else {
			go func() {
				stop <- status
			}()
		}

		select {
		case uuid := <-ch:
			asyncProcessStatusRequest.Uuid = uuid
			asyncProcessStatusRequest.OrgToken = os.Getenv("WS_APIKEY")

			jsonData, _ := asyncProcessStatusRequest.GetJsonData()
			body := AskProcessStatus(jsonData)
			json.Unmarshal(body, &asyncProcessResponse)
			if asyncProcessResponse.AsyncProcessStatus.Status == "SUCCESS" {
				status = "ok"
			} else {
				time.Sleep(5 * time.Second)
			}
		case <-stop:
			return asyncProcessResponse.AsyncProcessStatus.Status
		}
	}
}

func GetPrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer

	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}

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
		panic(err)
	}

	projectAlertRequest.InitRequest(updateRequestOrigin, uploadResponseData)
	jsonData, _ := projectAlertRequest.GetJsonData()
	body := AskProcessStatus(jsonData)

	return string(body)
}

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
		panic(err)
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
		panic(err)
	}
	defer resp.Body.Close()

	dPath := fmt.Sprintf(
		"report/%s/%s",
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
		panic(err)
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
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(body, &inventoryReport)
	if err != nil {
		panic(err)
	}

	return inventoryReport
}
