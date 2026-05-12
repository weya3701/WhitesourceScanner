package worker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

// TestReadURLsFromFile 測試 readURLsFromFile 函式。
func TestReadURLsFromFile(t *testing.T) {
	content := "http://example.com/file1.zip\n\nhttps://example.com/file2.tar.gz\ninvalid-url\nhttp://example.com/file3.exe"
	tmpFile, err := os.CreateTemp("", "urls-*.txt")
	if err != nil {
		t.Fatalf("無法建立暫存檔案: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("無法寫入暫存檔案: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("無法關閉暫存檔案: %v", err)
	}

	t.Run("讀取有效的URL", func(t *testing.T) {
		urls, err := readURLsFromFile(tmpFile.Name())
		if err != nil {
			t.Errorf("readURLsFromFile() 發生錯誤 = %v, 預期為 nil", err)
		}
		expected := []string{"http://example.com/file1.zip", "https://example.com/file2.tar.gz", "http://example.com/file3.exe"}
		if !reflect.DeepEqual(urls, expected) {
			t.Errorf("readURLsFromFile() 得到 = %v, 預期為 %v", urls, expected)
		}
	})

	t.Run("讀取不存在的檔案", func(t *testing.T) {
		_, err := readURLsFromFile("non-existent-file.txt")
		if err == nil {
			t.Error("readURLsFromFile() 讀取不存在的檔案時應回傳錯誤，但沒有")
		}
	})
}

// TestSetFilename 測試 DownloadTask 的 setFilename 方法。
func TestSetFilename(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{"包含檔名和副檔名的URL", "http://example.com/file.txt", "file.txt"},
		{"包含檔名但無副檔名的URL", "http://example.com/file", ""}, // 根據目前邏輯，只會取得包含'.'的檔名
		{"以斜線結尾的URL", "http://example.com/dir/", ""},
		{"只有域名的URL", "http://example.com", "example.com"},
		{"包含查詢參數的URL", "http://example.com/download.zip?file=data", "download.zip?file=data"},
		{"空URL", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dt := &DownloadTask{URL: tc.url}
			dt.setFilename()
			if dt.Filename != tc.expected {
				t.Errorf("setFilename() 對於 URL '%s': 得到 '%s', 預期 '%s'", tc.url, dt.Filename, tc.expected)
			}
		})
	}
}

// TestDownloadFile 測試 DownloadFile 函式。
func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/success" {
			fmt.Fprint(w, "file content")
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// 根據原始碼中的錯誤訊息，假設使用 curl 進行下載。
	t.Setenv("wget", "curl")

	tempDir, err := os.MkdirTemp("", "downloads")
	if err != nil {
		t.Fatalf("無法建立暫存目錄: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("成功下載", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 1)

		task := DownloadTask{
			URL:                 server.URL + "/success",
			Filename:            "testfile.txt",
			DownloadDestination: tempDir,
		}

		wg.Add(1)
		go DownloadFile(task, &wg, errChan)
		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				t.Errorf("DownloadFile() 回傳未預期的錯誤: %v", err)
			}
		}

		filePath := filepath.Join(tempDir, "testfile.txt")
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("無法讀取下載的檔案: %v", err)
		}
		if string(content) != "file content" {
			t.Errorf("下載的檔案內容為 = %s, 預期為 = 'file content'", string(content))
		}
	})

	t.Run("下載失敗 (404 Not Found)", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 1)

		task := DownloadTask{
			URL:                 server.URL + "/notfound",
			Filename:            "notfound.txt",
			DownloadDestination: tempDir,
		}

		wg.Add(1)
		go DownloadFile(task, &wg, errChan)
		wg.Wait()
		close(errChan)

		err, ok := <-errChan
		if !ok || err == nil {
			t.Error("DownloadFile() 使用 404 URL 時應回傳錯誤，但沒有")
		}
	})
}

// TestParallelDownload 測試 ParallelDownload 函式。
func TestParallelDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "content for %s", r.URL.Path)
	}))
	defer server.Close()

	t.Setenv("wget", "curl")

	tempDir, err := os.MkdirTemp("", "parallel_downloads")
	if err != nil {
		t.Fatalf("無法建立暫存目錄: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("所有下載成功", func(t *testing.T) {
		tasks := []DownloadTask{
			{URL: server.URL + "/file1", Filename: "file1.txt", DownloadDestination: tempDir},
			{URL: server.URL + "/file2", Filename: "file2.txt", DownloadDestination: tempDir},
			{URL: server.URL + "/file3", Filename: "file3.txt", DownloadDestination: tempDir},
		}

		err := ParallelDownload(tasks, 2)
		if err != nil {
			t.Errorf("ParallelDownload() 回傳未預期的錯誤: %v", err)
		}

		for _, task := range tasks {
			filePath := filepath.Join(task.DownloadDestination, task.Filename)
			content, readErr := os.ReadFile(filePath)
			if readErr != nil {
				t.Fatalf("無法讀取下載的檔案 %s: %v", task.Filename, readErr)
			}
			expectedContent := fmt.Sprintf("content for /%s", filepath.Base(task.URL))
			if string(content) != expectedContent {
				t.Errorf("檔案 %s 內容不匹配: 得到 %q, 預期 %q", task.Filename, string(content), expectedContent)
			}
		}
	})

	t.Run("其中一個下載失敗", func(t *testing.T) {
		tasks := []DownloadTask{
			{URL: server.URL + "/file4", Filename: "file4.txt", DownloadDestination: tempDir},
			{URL: server.URL + "/fail", Filename: "fail.txt", DownloadDestination: tempDir},
			{URL: server.URL + "/file5", Filename: "file5.txt", DownloadDestination: tempDir},
		}

		err := ParallelDownload(tasks, 2)
		if err == nil {
			t.Error("ParallelDownload() 應回傳錯誤，但沒有")
		}
	})
}
