package main

import (
	"artifact_repository/handler"
	"flag"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	mode := flag.String("mode", "", "App Mode")
	packageName := flag.String("package_name", "", "Package Name")
	projectName := flag.String("project_name", "", "Project Name")
	withConf := flag.String("with_conf", "", "With Config")
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
	}
}
