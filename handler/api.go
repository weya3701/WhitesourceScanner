package handler

import (
	"artifact_repository/worker"
	"artifact_repository/wss"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetTemplate(tpl string) string {
	return fmt.Sprintf("%s.tmpl", tpl)
}

func PackagePage(c *gin.Context) {
	c.HTML(http.StatusOK, GetTemplate("package"), nil)
}

func SyncPage(c *gin.Context) {
	c.HTML(http.StatusOK, GetTemplate("sync"), nil)
}

func GetPackage(c *gin.Context) {

	var (
		packageName    string
		packageType    string
		packageVersion string
		withConf       string = ""
	)
	var wk worker.WorkerHandler
	InitPage(c, packageName, packageType, packageVersion)
	projectName := DownloadPackage(
		packageName,
		packageType,
		packageVersion,
		os.Getenv("internet_index"),
		"./tmp/",
	)

	wss.DoWhitesourceScan(packageName, projectName, withConf)
	wss.DoUploadRequest(projectName)

	ch := wss.GenerateProjectReportAsync(projectName)
	_ = wss.GetProcessStatus(ch, projectName)

	reportPath := fmt.Sprintf("report/%s", projectName)
	os.Mkdir(reportPath, 0755)
	rsp := wss.GetProjectRiskReport(projectName)
	_, err := json.Marshal(rsp)
	if err != nil {
		panic(err)
	}
	worker.UploadToRepository(wk, os.Getenv("tmp_api"), fmt.Sprintf("./tmp/%s", packageName))
	c.HTML(http.StatusOK, GetTemplate("package"), gin.H{
		"report":      fmt.Sprintf("http://localhost:8888/report/%s/risk.pdf", strings.Split(packageName, "==")[0]),
		"projectName": projectName,
		"sync_url":    "http://localhost:8888/sync",
	})
}

func InitPage(c *gin.Context, packageName, packageType, packageVersion string) {
	if in, isExist := c.GetPostForm("packageName"); isExist && in != "" {
		packageName = in
	} else {
		c.HTML(http.StatusBadRequest, GetTemplate("package"), gin.H{
			"error": errors.New("Please input package name"),
		})
	}
	if in, isExist := c.GetPostForm("packageType"); isExist && in != "" {
		packageType = in
	} else {
		c.HTML(http.StatusBadRequest, GetTemplate("package"), gin.H{
			"error": errors.New("Please input package type"),
		})
	}
	if in, isExist := c.GetPostForm("packageVersion"); isExist && in != "" {
		packageVersion = in
	} else {
		c.HTML(http.StatusBadRequest, "packageVersion", gin.H{
			"error": errors.New("Please input package version"),
		})
	}
}

func DownloadPackage(packageName, packageType, packageVersion, indexUrl, destination string) string {

	var wk worker.WorkerHandler
	packageName = fmt.Sprintf("%s==%s", packageName, packageVersion)
	if packageType == "pypi" {
		wk = worker.NewRepositoryWorker(worker.Pypi{})
	}
	wk.DownloadFromIndex(destination, packageName, indexUrl)
	return strings.Split(packageName, "==")[0]
}

func SyncPackage(c *gin.Context) {

	var (
		packageName    string
		packageType    string
		packageVersion string
	)
	var wk worker.WorkerHandler
	InitPage(c, packageName, packageType, packageVersion)

	DownloadPackage(
		packageName,
		packageType,
		packageVersion,
		os.Getenv("tmp_index"),
		"./sync_tmp/",
	)

	worker.UploadToRepository(wk, os.Getenv("prod_api"), fmt.Sprintf("./tmp/%s", packageName))
	c.HTML(http.StatusOK, GetTemplate("sync"), nil)

}
