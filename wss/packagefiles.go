package wss

import (
	"fmt"
	"os"
)

func CreateDir(filename string, mode os.FileMode) int {
	package_tmp_folder := os.Getenv("package_tmp")

	filename = fmt.Sprintf("%s/%s", package_tmp_folder, filename)
	err := os.MkdirAll(filename, mode)

	if err != nil {
		return 0
	}

	return 1
}
