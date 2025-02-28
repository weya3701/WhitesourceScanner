package worker

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type Worker interface {
	Download(destination string, packageName string, indexUrl string) string
	Sync(targetUrl string, packageFile string) string
	Remove(fullPath string) error
}

type WorkerHandler struct {
	worker Worker
}

func (rw WorkerHandler) Download(packageName string, indexUrl string) {
	rw.worker.Download(
		os.Getenv("package_tmp"),
		packageName,
		indexUrl,
	)
}

func (rw WorkerHandler) DownloadFromIndex(destination string, packageName string, indexUrl string) {
	rw.worker.Download(
		destination,
		packageName,
		indexUrl,
	)
}

func (rw WorkerHandler) Sync(targetUrl string, packageFile string) {
	rw.worker.Sync(
		targetUrl,
		packageFile,
	)
}

func (rw WorkerHandler) Remove(fullPath string) error {
	return rw.worker.Remove(fullPath)
}

func NewRepositoryWorker(worker Worker) WorkerHandler {
	return WorkerHandler{worker: worker}
}

func UploadToRepository(worker WorkerHandler, targetUrl string, sourcePath string) {

	var wg sync.WaitGroup

	files, err := ioutil.ReadDir(sourcePath)
	if err != nil {
		fmt.Println("讀取目錄錯誤: ", err)
		panic(err)
	}
	for _, file := range files {
		wg.Add(1)
		pkgName := fmt.Sprintf("%s/%s", sourcePath, file.Name())
		go func() {
			worker.Sync(targetUrl, pkgName)
			wg.Done()
		}()
	}
	wg.Wait()
}
