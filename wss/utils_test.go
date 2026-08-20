package wss

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func useTestHTTPClient(t *testing.T, status int, body string) {
	t.Helper()
	previous := apiHTTPClient
	apiHTTPClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", request.Method)
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})}
	t.Cleanup(func() { apiHTTPClient = previous })
}

func TestAskProcessStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "success", status: http.StatusOK},
		{name: "server error", status: http.StatusInternalServerError, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useTestHTTPClient(t, tt.status, `{"asyncProcessStatus":{"status":"SUCCESS"}}`)
			t.Setenv("whitesource_api", "https://example.test/api")
			err, _ := AskProcessStatus([]byte(`{}`))
			if (err != nil) != tt.wantErr {
				t.Fatalf("AskProcessStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAskProcessStatusVerboseOutput(t *testing.T) {
	output := captureVerboseOutput(t)
	SetVerbose(true)
	useTestHTTPClient(t, http.StatusOK, `{"asyncProcessStatus":{"status":"SUCCESS"}}`)
	t.Setenv("whitesource_api", "https://example.test/api")

	err, _ := AskProcessStatus([]byte(`{"requestType":"getAsyncProcessStatus"}`))
	if err != nil {
		t.Fatalf("AskProcessStatus() error = %v", err)
	}
	got := output.String()
	for _, want := range []string{"request_type=getAsyncProcessStatus", "status_code=200", "bytes="} {
		if !strings.Contains(got, want) {
			t.Fatalf("verbose output = %q, want %q", got, want)
		}
	}
}

func TestGetProcessStatusReturnsFailure(t *testing.T) {
	dir := t.TempDir()
	project := "project"
	projectDir := filepath.Join(dir, project)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	request := UpdateRequestOriginal{UserKey: "user"}
	requestJSON, _ := json.Marshal(request)
	if err := os.WriteFile(filepath.Join(projectDir, "request.json"), requestJSON, 0644); err != nil {
		t.Fatal(err)
	}
	data := UploadResponseData{ProjectNamesToDetails: map[string]ProjectInfo{"project": {ProjectToken: "token"}}}
	dataJSON, _ := json.Marshal(data)
	statusJSON, _ := json.Marshal(UploadResponseStatus{Data: string(dataJSON)})
	if err := os.WriteFile(filepath.Join(projectDir, "status.json"), statusJSON, 0644); err != nil {
		t.Fatal(err)
	}

	useTestHTTPClient(t, http.StatusOK, `{"asyncProcessStatus":{"status":"FAILED"}}`)
	t.Setenv("whitesource_path", dir)
	t.Setenv("request_file", "request.json")
	t.Setenv("response_status_file", "status.json")
	t.Setenv("whitesource_api", "https://example.test/api")
	status, err := GetProcessStatus("uuid", project)
	if err == nil || status != "FAILED" {
		t.Fatalf("GetProcessStatus() = (%q, %v), want FAILED error", status, err)
	}
}

func TestUploadResponseToFileTruncates(t *testing.T) {
	file := filepath.Join(t.TempDir(), "status.json")
	if err := os.WriteFile(file, []byte(strings.Repeat("x", 1024)), 0644); err != nil {
		t.Fatal(err)
	}
	want := UploadResponseStatus{Status: 1}
	if !want.ToFile(file) {
		t.Fatal("ToFile returned false")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var got UploadResponseStatus
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("written JSON is invalid: %v", err)
	}
	if got.Status != want.Status {
		t.Fatalf("status = %d, want %d", got.Status, want.Status)
	}
}
