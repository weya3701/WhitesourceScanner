package wss

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetJsonContentType(t *testing.T) {
	key, value := GetJsonContentType()
	if key != "Content-Type" {
		t.Errorf("Expected key 'Content-Type', got '%s'", key)
	}
	if value != "application/json" {
		t.Errorf("Expected value 'application/json', got '%s'", value)
	}
}

func TestGetFilePath(t *testing.T) {
	path := "/tmp/ws/"
	projectName := "my-project"
	fileName := "update-request.json"
	expected := "/tmp/ws/my-project/update-request.json"
	result := GetFilePath(path, projectName, fileName)
	if result != expected {
		t.Errorf("GetFilePath() = %q, want %q", result, expected)
	}
}

func TestGetPrettyString(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:  "Valid JSON",
			input: `{"a":1,"b":"hello"}`,
			expected: `{
    "a": 1,
    "b": "hello"
}`,
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			input:       `{"a":1,"b":"hello"`,
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetPrettyString(tc.input)
			if (err != nil) != tc.expectError {
				t.Fatalf("GetPrettyString() error = %v, expectError %v", err, tc.expectError)
			}
			if !tc.expectError && result != tc.expected {
				t.Errorf("GetPrettyString() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestAskProcessStatus(t *testing.T) {
	expectedBody := `{"status": "SUCCESS"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected 'POST' request, got '%s'", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Charset") != "utf-8" {
			t.Errorf("Expected Charset 'utf-8', got '%s'", r.Header.Get("Charset"))
		}
		fmt.Fprintln(w, expectedBody)
	}))
	defer server.Close()

	originalApi := os.Getenv("whitesource_api")
	os.Setenv("whitesource_api", server.URL)
	defer os.Setenv("whitesource_api", originalApi)

	jsonData := []byte(`{"request": "test"}`)
	body := AskProcessStatus(jsonData)

	// httptest server adds a newline, so we need to trim it.
	if string(bytes.TrimSpace(body)) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(body))
	}
}
