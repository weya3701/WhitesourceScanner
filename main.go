package main

import (
	"flag"
	"io"
	"os"
	"wss/handler"
	"wss/wss"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// ExportFile  string
	// Application string
	// TarFile     string
	// ImageName   string
	// imageTag    string

	mode := flag.String("mode", "", "App Mode")
	packageName := flag.String("package_name", "", "Package Name")
	projectName := flag.String("project_name", "", "Project Name")
	withConf := flag.String("with_conf", "", "With Config")
	exportFile := flag.String("export_file", "", "Export File")
	application := flag.String("application", "", "Application")
	tarFile := flag.String("tar_file", "", "Tar File")
	imageName := flag.String("image_name", "", "Image Name")
	imageTag := flag.String("image_tag", "", "Image Tag")
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	if *mode == "api" {
		log_file, _ := os.Create("./logs/gin.log")
		gin.DefaultWriter = io.MultiWriter(log_file)
		server := gin.Default()

		server.LoadHTMLGlob("template/html/*")
		server.Static("/assets", "./template/assets")
		server.Static("/report", "./report/")
		server.GET("/package", handler.PackagePage)
		server.GET("/sync", handler.SyncPage)
		server.POST("/get_package", handler.GetPackage)
		server.POST("/sync_package", handler.SyncPackage)

		server.Run(":8888")

	}
	if *mode == "cmd" {
		handler.GetPackageReport(*packageName, *projectName, *withConf)
		handler.GetProjectAlert(*projectName)
		handler.UpdateRiskReport(*projectName)
		handler.GetInventoryReport(*projectName)
	}
	if *mode == "image" {
		mendCli := handler.InitMendCli(
			*exportFile,
			*application,
			*packageName,
			*projectName,
			*tarFile,
			*imageName,
			*imageTag,
		)
		wss.DoMendCliScan(mendCli)
	}
}
