package wss

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RuntimeConfig centralizes process-level configuration read from the environment.
type RuntimeConfig struct {
	SettingsFile       string
	PackageTmp         string
	ReportTmp          string
	WhiteSourcePath    string
	WhiteSourceAgent   string
	WhiteSourceAPI     string
	RequestFile        string
	ResponseStatusFile string
	ResponseDataFile   string
	RiskReportFile     string
	AgentPath          string
	AgentName          string
	AgentURL           string
	WgetCommand        string
	PipCommand         string
	MavenCommand       string
	NPMCommand         string
	GradleCommand      string
	Concurrency        int
}

func LoadRuntimeConfig() (RuntimeConfig, error) {
	concurrency := 4
	if value := strings.TrimSpace(os.Getenv("concurrency")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return RuntimeConfig{}, fmt.Errorf("concurrency must be a positive integer")
		}
		concurrency = parsed
	}
	return RuntimeConfig{
		SettingsFile: os.Getenv("settings_file"), PackageTmp: os.Getenv("package_tmp"),
		ReportTmp: os.Getenv("report_tmp"), WhiteSourcePath: os.Getenv("whitesource_path"),
		WhiteSourceAgent: os.Getenv("whitesource_agent"), WhiteSourceAPI: os.Getenv("whitesource_api"),
		RequestFile: os.Getenv("request_file"), ResponseStatusFile: os.Getenv("response_status_file"),
		ResponseDataFile: os.Getenv("response_data_file"), RiskReportFile: os.Getenv("risk_report_file"),
		AgentPath: os.Getenv("wssAgentPath"), AgentName: os.Getenv("wssAgentName"), AgentURL: os.Getenv("agentURL"),
		WgetCommand: os.Getenv("wget"), PipCommand: os.Getenv("pip"), MavenCommand: os.Getenv("maven"),
		NPMCommand: os.Getenv("npm"), GradleCommand: os.Getenv("gradle"), Concurrency: concurrency,
	}, nil
}

func (c RuntimeConfig) Validate(mode, packageType string) error {
	if mode == "image" {
		return nil
	}
	required := map[string]string{
		"settings_file": c.SettingsFile, "package_tmp": c.PackageTmp, "report_tmp": c.ReportTmp,
		"whitesource_path": c.WhiteSourcePath, "whitesource_agent": c.WhiteSourceAgent,
		"whitesource_api": c.WhiteSourceAPI, "request_file": c.RequestFile,
		"response_status_file": c.ResponseStatusFile, "response_data_file": c.ResponseDataFile,
		"risk_report_file": c.RiskReportFile, "wssAgentPath": c.AgentPath, "wssAgentName": c.AgentName,
	}
	if packageType == "pip" {
		required["pip"] = c.PipCommand
	} else if packageType == "maven" {
		required["maven"] = c.MavenCommand
	} else if packageType == "npm" {
		required["npm"] = c.NPMCommand
	} else if packageType == "gradle" {
		required["gradle"] = c.GradleCommand
	} else if packageType == "wget" {
		required["wget"] = c.WgetCommand
	}
	var missing []string
	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	return nil
}
