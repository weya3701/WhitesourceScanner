package main

import (
	"flag"
	"fmt"
	"wss/handler"
	"wss/wss"

	"github.com/joho/godotenv"
)

type BatchFunc func() (bool, error)

type BatchTask struct {
	Name string
	Func BatchFunc
}

type BatchRunner struct {
	tasks []BatchTask
}

func NewBatchRunner(tasks []BatchTask) *BatchRunner {
	return &BatchRunner{
		tasks: tasks,
	}
}

func (br *BatchRunner) Run() (bool, error) {
	fmt.Println("開始批次執行...")
	for i, task := range br.tasks {
		fmt.Printf("--- 正在執行任務: %s (序號: %d) ---\n", task.Name, i+1)

		success, err := task.Func()
		if err != nil {
			fmt.Printf("任務 '%s' 執行失敗，錯誤: %v\n", task.Name, err)
			return false, fmt.Errorf("任務 '%s' 執行失敗: %w", task.Name, err)
		}
		if !success {
			fmt.Printf("任務 '%s' 執行結果為失敗，停止批次執行。\n", task.Name)
			return false, fmt.Errorf("任務 '%s' 返回失敗狀態，批次中止", task.Name)
		}

		fmt.Printf("任務 '%s' 執行成功。\n\n", task.Name)
	}

	fmt.Println("所有批次任務執行完成。")
	return true, nil
}

func main() {

	mode := flag.String("mode", "", "App Mode")
	packageName := flag.String("package_name", "", "Package Name")
	projectName := flag.String("project_name", "", "Project Name")
	withConf := flag.String("with_conf", "", "With Config")
	exportFile := flag.String("export_file", "", "Export File")
	application := flag.String("application", "", "Application")
	tarFile := flag.String("tar_file", "", "Tar File")
	imageName := flag.String("image_name", "", "Image Name")
	imageTag := flag.String("image_tag", "", "Image Tag")
	packageType := flag.String("package_type", "", "Package Type")
	requirementsFile := flag.String("requirements_file", "", "Requirements File")
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load environ")
	}

	if *mode == "reqfile" {
		tasks := []BatchTask{
			{
				Name: "同步定義套件",
				Func: func() (bool, error) {
					handler.SyncDefinitionPackages(*packageType, *projectName, *requirementsFile)
					return true, nil
				},
			},
			{
				Name: "取得套件報告",
				Func: func() (bool, error) {
					return handler.GetPackageReport(*packageName, *projectName, *withConf)
				},
			},
			{
				Name: "取得專案警報",
				Func: func() (bool, error) {
					return handler.GetProjectAlert(*projectName)
				},
			},
			{
				Name: "更新風險報告",
				Func: func() (bool, error) {
					err := handler.UpdateRiskReport(*projectName)
					if err != nil {
						return false, err
					}
					return true, nil
				},
			},
			{
				Name: "取得庫存報告",
				Func: func() (bool, error) {
					return handler.GetInventoryReport(*projectName, *packageType)
				},
			},
		}

		runner := NewBatchRunner(tasks)
		if success, runErr := runner.Run(); !success {
			fmt.Printf("批次執行失敗: %v\n", runErr)
		} else {
			fmt.Println("批次執行成功完成！")
		}
	}
	if *mode == "cmd" {
		tasks := []BatchTask{
			{
				Name: "取得套件報告",
				Func: func() (bool, error) {
					return handler.GetPackageReport(*packageName, *projectName, *withConf)
				},
			},
			{
				Name: "取得專案警報",
				Func: func() (bool, error) {
					return handler.GetProjectAlert(*projectName)
				},
			},
			{
				Name: "更新風險報告",
				Func: func() (bool, error) {
					err := handler.UpdateRiskReport(*projectName)
					if err != nil {
						return false, err
					}
					return true, nil
				},
			},
			{
				Name: "取得庫存報告",
				Func: func() (bool, error) {
					handler.GetInventoryReport(*projectName, *packageType)
					return true, nil
				},
			},
		}

		runner := NewBatchRunner(tasks)
		if success, runErr := runner.Run(); !success {
			fmt.Printf("批次執行失敗: %v\n", runErr)
		} else {
			fmt.Println("批次執行成功完成！")
		}
	}
	if *mode == "image" {
		tasks := []BatchTask{
			{
				Name: "執行 Docker Tar 檔案掃描",
				Func: func() (bool, error) {
					mendCli := handler.InitMendCli(
						*exportFile,
						*application,
						*packageName,
						*projectName,
						*tarFile,
						*imageName,
						*imageTag,
					)
					wss.DoDockerTarFileScan(mendCli)
					return true, nil
				},
			},
		}

		runner := NewBatchRunner(tasks)
		if success, runErr := runner.Run(); !success {
			fmt.Printf("批次執行失敗: %v\n", runErr)
		} else {
			fmt.Println("批次執行成功完成！")
		}
	}
}
