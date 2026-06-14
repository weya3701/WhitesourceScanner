package wss

import (
	"fmt"
	"os"
)

// CreateDir 在 package_tmp 環境變數指定的基礎路徑下建立一個目錄。
// 如果目錄已存在，則不會返回錯誤。
//
// 參數:
//   - filename: 要建立的目錄名稱。
//   - mode: 目錄的檔案權限模式，例如 0755。
//
// 返回:
//   - int: 如果目錄成功建立或已存在，返回 1；如果創建失敗，返回 0。
func CreateDir(filename string, mode os.FileMode) int {
	package_tmp_folder := os.Getenv("package_tmp")

	filename = fmt.Sprintf("%s/%s", package_tmp_folder, filename)
	err := os.MkdirAll(filename, mode)

	if err != nil {
		return 0
	}

	return 1
}
