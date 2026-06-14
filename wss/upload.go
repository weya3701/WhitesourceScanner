package wss

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// NewUpdateRequestFromFile 從指定檔案路徑讀取並解析 UpdateRequestOriginal 結構。
//
// 參數:
//   - filepath: 請求檔案的路徑。
//
// 返回:
//   - UpdateRequestOriginal: 解析後的 UpdateRequestOriginal 結構實例。
func NewUpdateRequestFromFile(filepath string) UpdateRequestOriginal {
	var updateRequestOrigin UpdateRequestOriginal
	data, err := os.ReadFile(filepath)
	err = json.Unmarshal(data, &updateRequestOrigin)
	if err != nil {
		log.Printf("Json Unmarshal failed: %s", err)
	}
	return updateRequestOrigin
}

// GetValues 將 UpdateRequestOriginal 結構轉換為 url.Values 格式。
// 它會將結構中的欄位（如 updateType, type, agent 等）以及 diff 資料序列化為 URL 查詢參數。
//
// 返回:
//   - url.Values: 包含請求資料的 URL 查詢參數。
func (u UpdateRequestOriginal) GetValues() url.Values {
	diff_data, err := json.Marshal(u.Diff)
	if err != nil {

		fmt.Printf("Json Marshal failed: %s", err)
	}
	values := url.Values{}
	values.Set("updateType", u.UpdateType)
	values.Set("type", u.Type)
	values.Set("agent", u.Agent)
	values.Set("agentVersion", u.AgentVersion)
	values.Set("token", u.Token)
	values.Set("userKey", u.UserKey)
	values.Set("timeStamp", strconv.Itoa(u.TimeStamp))
	values.Set("product", u.Product)
	values.Set("diff", string(diff_data))

	return values
}

// LoadUpdateRequest 從指定檔案路徑載入並解析 UpdateRequestOriginal 結構。
// 如果讀取檔案或 JSON 解析失敗，將會記錄錯誤。
//
// 參數:
//   - filepath: 請求檔案的路徑。
func (u *UpdateRequestOriginal) LoadUpdateRequest(filepath string) {

	data, err := os.ReadFile(filepath)
	if err != nil {

		log.Printf("Read file failed: %s", err)
		return
	}
	err = json.Unmarshal(data, &u)
	if err != nil {
		log.Printf("Json Unmarshal failed: %s", err)
		return
	}
}

// SendUploadRequest 向指定的 WSS URL 發送上傳請求。
// 請求的內容是 UpdateRequestOriginal 結構序列化後的 URL 查詢參數。
//
// 參數:
//   - wssurl: 白源安全 (WSS) 服務的 URL。
//
// 返回:
//   - *http.Response: 伺服器回應。
//   - error: 如果發送請求失敗，返回錯誤。
func (u UpdateRequestOriginal) SendUploadRequest(wssurl string) (resp *http.Response, err error) {

	vals := u.GetValues()
	req, err := http.NewRequest("POST", wssurl, bytes.NewBuffer([]byte(vals.Encode())))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Add("Accept-Charset", "utf-8")
	res, err := http.DefaultClient.Do(req)
	return res, err
}

// FromFile 從指定檔案路徑讀取並解析 UpdateRequestOriginal 結構。
//
// 參數:
//   - fromfile: 請求檔案的路徑。
//
// 返回:
//   - bool: 如果成功解析，返回 true；否則返回 false。
func (u *UpdateRequestOriginal) FromFile(fromfile string) bool {
	var status bool = true

	bytes, err := os.ReadFile(fromfile)
	if err != nil {
		status = false
	}

	err = json.Unmarshal(bytes, &u)
	if err != nil {
		status = false
	}

	return status
}

// GetProjectName 從 UploadResponseData 中獲取第一個專案的名稱。
//
// 返回:
//   - string: 專案名稱。
func (ud UploadResponseData) GetProjectName() string {
	var projectName string = ""

	for k := range ud.ProjectNamesToDetails {
		projectName = k
	}

	return projectName
}

// GetJson 將 UploadResponseStatus 結構序列化為 JSON 格式的位元組陣列。
//
// 返回:
//   - []byte: JSON 格式的位元組陣列。
func (us UploadResponseStatus) GetJson() []byte {
	var data []byte
	data, err := json.Marshal(us)

	if err != nil {
		log.Println("Failed to json marshal")
	}
	return data
}

// GetJson 將 UploadResponseData 結構序列化為 JSON 格式的位元組陣列。
//
// 返回:
//   - []byte: JSON 格式的位元組陣列。
func (ud UploadResponseData) GetJson() []byte {
	var data []byte
	data, err := json.Marshal(ud)

	if err != nil {
		log.Println("Failed to json marshal")
	}
	return data
}

// ToFile 將 UploadResponseStatus 結構的 JSON 內容寫入指定檔案。
//
// 參數:
//   - destination: 輸出檔案的路徑。
//
// 返回:
//   - bool: 如果成功寫入，返回 true；否則返回 false。
func (us UploadResponseStatus) ToFile(destination string) bool {
	var status bool = true

	file, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		status = false
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	data := us.GetJson()
	_, err = writer.Write(data)
	if err != nil {
		status = false
	}
	err = writer.Flush()
	if err != nil {
		log.Println("Write flush failed")
		status = false
	}

	return status
}

// ToFile 將 UploadResponseData 結構的 JSON 內容寫入指定檔案。
//
// 參數:
//   - destination: 輸出檔案的路徑。
//
// 返回:
//   - bool: 如果成功寫入，返回 true；否則返回 false。
func (ud UploadResponseData) ToFile(destination string) bool {
	var status bool = true

	file, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, os.FileMode(0644))

	if err != nil {
		status = false
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	data := ud.GetJson()
	_, err = writer.Write(data)
	if err != nil {
		status = false
	}

	err = writer.Flush()
	if err != nil {
		log.Println("Write flush failed")
		status = false
	}
	return status
}

// FromFile 從指定檔案路徑讀取並解析 UploadResponseData 結構。
//
// 參數:
//   - fromfile: 檔案的路徑。
//
// 返回:
//   - bool: 如果成功解析，返回 true；否則返回 false。
func (ud *UploadResponseData) FromFile(fromfile string) bool {
	var status bool = true

	bytes, err := os.ReadFile(fromfile)
	if err != nil {
		status = false
		fmt.Println("Read file failed")
	}
	err = json.Unmarshal(bytes, &ud)
	if err != nil {
		status = false
		log.Println("json unmarshal failed")
	}

	return status

}

// FromFile 從指定檔案路徑讀取並解析 UploadResponseStatus 結構。
//
// 參數:
//   - fromfile: 檔案的路徑。
//
// 返回:
//   - bool: 如果成功解析，返回 true；否則返回 false。
func (us *UploadResponseStatus) FromFile(fromfile string) bool {
	var status bool = true

	bytes, err := os.ReadFile(fromfile)
	if err != nil {
		status = false
		fmt.Println("Read file failed")
	}

	err = json.Unmarshal(bytes, &us)
	if err != nil {
		status = false
		log.Println("json unmarshal failed")
	}

	return status
}
