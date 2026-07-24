package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"wss/handler"
	"wss/wss"
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
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "執行失敗: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	mode := flag.String("mode", "", "App Mode")
	packageName := flag.String("package_name", "", "Package Name")
	projectName := flag.String("project_name", "", "Project Name")
	scanSource := flag.String("scan_source", "", "Scan source path (cmd mode only)")
	withConf := flag.String("with_conf", "", "With Config")
	exportFile := flag.String("export_file", "", "Export File")
	application := flag.String("application", "", "Application")
	tarFile := flag.String("tar_file", "", "Tar File")
	imageName := flag.String("image_name", "", "Image Name")
	imageTag := flag.String("image_tag", "", "Image Tag")
	packageType := flag.String("package_type", "", "Package Type")
	requirementsFile := flag.String("requirements_file", "", "Requirements File")
	flag.Parse()

	if err := wss.LoadEnvironment(".env"); err != nil {
		return err
	}

	if err := validateArguments(*mode, *packageName, *projectName, *scanSource, *application, *packageType, *requirementsFile); err != nil {
		return err
	}
	effectivePackageName := *packageName
	effectiveProjectName := *projectName
	directScanSource := false
	if *mode == "cmd" && strings.TrimSpace(*scanSource) != "" {
		effectivePackageName = filepath.Clean(*scanSource)
		effectiveProjectName = filepath.Base(effectivePackageName)
		directScanSource = true
	}
	config, err := wss.LoadRuntimeConfig()
	if err != nil {
		return err
	}
	if err := config.Validate(*mode, *packageType); err != nil {
		return err
	}

	var tasks []BatchTask
	switch *mode {
	case "reqfile":
		tasks = []BatchTask{
			{
				Name: "同步定義套件",
				Func: func() (bool, error) {
					err := handler.SyncDefinitionPackages(*packageType, effectiveProjectName, *requirementsFile)
					return err == nil, err
				},
			},
			{
				Name: "取得套件報告",
				Func: func() (bool, error) {
					return handler.GetPackageReport(effectivePackageName, effectiveProjectName, *withConf, false)
				},
			},
			{
				Name: "取得專案警報",
				Func: func() (bool, error) {
					return handler.GetProjectAlert(effectiveProjectName)
				},
			},
			{
				Name: "取得庫存報告",
				Func: func() (bool, error) {
					return handler.GetInventoryReport(effectiveProjectName, *packageType)
				},
			},
		}

	case "cmd":
		tasks = []BatchTask{
			{
				Name: "取得套件報告",
				Func: func() (bool, error) {
					return handler.GetPackageReport(effectivePackageName, effectiveProjectName, *withConf, directScanSource)
				},
			},
			{
				Name: "取得專案警報",
				Func: func() (bool, error) {
					return handler.GetProjectAlert(effectiveProjectName)
				},
			},
			{
				Name: "取得庫存報告",
				Func: func() (bool, error) {
					return handler.GetInventoryReport(effectiveProjectName, *packageType)
				},
			},
		}
	case "image":
		tasks = []BatchTask{
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
					err := wss.DoDockerTarFileScan(mendCli)
					return err == nil, err
				},
			},
		}
	}

	if *mode == "cmd" || *mode == "reqfile" {
		archiveSource := filepath.Join(os.Getenv("package_tmp"), effectivePackageName)
		if directScanSource {
			archiveSource = effectivePackageName
		}
		tasks = append(tasks, BatchTask{
			Name: "零風險掃描來源封裝",
			Func: func() (bool, error) {
				archiveName, err := wss.ArchiveReportIfNoVulnerabilities(
					os.Getenv("report_tmp"),
					effectiveProjectName,
					archiveSource,
					".",
				)
				if err == nil && archiveName != "" {
					fmt.Printf("報告已封裝為 %s\n", archiveName)
				}
				return err == nil, err
			},
		})
	}

	runner := NewBatchRunner(tasks)
	if success, err := runner.Run(); !success {
		return err
	}
	fmt.Println("批次執行成功完成！")
	return nil
}

func validateArguments(mode, packageName, projectName, scanSource, application, packageType, requirementsFile string) error {
	if mode != "cmd" && mode != "reqfile" && mode != "image" {
		return fmt.Errorf("mode 必須是 cmd、reqfile 或 image")
	}
	if scanSource != "" {
		if mode != "cmd" {
			return fmt.Errorf("scan_source 僅適用於 cmd 模式")
		}
		cleanSource := filepath.Clean(scanSource)
		sourceName := filepath.Base(cleanSource)
		if sourceName == "." || sourceName == string(filepath.Separator) {
			return fmt.Errorf("scan_source 必須包含可作為專案名稱的目錄名稱")
		}
		info, err := os.Stat(cleanSource)
		if err != nil {
			return fmt.Errorf("無法讀取 scan_source: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("scan_source 必須是目錄")
		}
		return nil
	}
	if projectName == "" {
		return fmt.Errorf("project_name 為必填")
	}
	if strings.ContainsAny(projectName, `/\\`) || projectName == "." || projectName == ".." {
		return fmt.Errorf("project_name 不可包含路徑分隔符")
	}
	if mode == "image" {
		if application == "" {
			return fmt.Errorf("image 模式需要 application")
		}
		return nil
	}
	if packageName == "" {
		return fmt.Errorf("%s 模式需要 package_name", mode)
	}
	if mode == "reqfile" {
		if packageType == "" || requirementsFile == "" {
			return fmt.Errorf("reqfile 模式需要 package_type 與 requirements_file")
		}
		supported := map[string]bool{"pip": true, "maven": true, "npm": true, "gradle": true, "wget": true}
		if !supported[packageType] {
			return fmt.Errorf("package_type 必須是 pip、maven、npm、gradle 或 wget")
		}
		info, err := os.Stat(requirementsFile)
		if err != nil {
			return fmt.Errorf("無法讀取 requirements_file: %w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("requirements_file 不可為目錄")
		}
	}
	return nil
}
