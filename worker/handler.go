package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Worker interface {
	Download(destination string, packageName string, indexUrl string) string
	Sync(targetUrl string, packageFile string) string
	Remove(fullPath string) error
	SyncPackages(destination string, requirementsFile string) error
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

func (rw WorkerHandler) SyncPackagesFromDefintionFile(projectName string, requirementsFile string) {
	rw.worker.SyncPackages(projectName, requirementsFile)
}

func (rw WorkerHandler) Remove(fullPath string) error {
	return rw.worker.Remove(fullPath)
}

func NewRepositoryWorker(worker Worker) WorkerHandler {
	return WorkerHandler{worker: worker}
}

func UploadToRepository(worker WorkerHandler, targetUrl string, sourcePath string) {

	var wg sync.WaitGroup

	// files, err := ioutil.ReadDir(sourcePath)
	files, err := os.ReadDir(sourcePath)
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

func GetDependenciesTree(filename string, prefix string, cmds []string) error {

	var err error = nil

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, prefix, cmds...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gradle dependencies failed: %w, output %s", err, string(out))
	}
	fmt.Println(string(out))

	err = os.WriteFile(filename, out, 0644) // 0644 是檔案權限，可根據需要調整
	if err != nil {
		return fmt.Errorf("failed to write output to file %s: %w", filename, err)
	}

	return err
}
