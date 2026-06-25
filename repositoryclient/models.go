package repositoryclient

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3" // 導入 yaml 庫
)

// RepositoryType 定義了儲存庫的類型。
// type RepositoryType string
//
// const (
// 	RepositoryTypeNPM    RepositoryType = "npm"
// 	RepositoryTypeGradle RepositoryType = "gradle"
// 	RepositoryTypeMaven  RepositoryType = "maven"
// )

// RepositoryConnection 包含了連接到儲存庫並發佈 Artifacts 所需的所有資訊。
type RepositoryConnection struct {
	Type      string `yaml:"type"`      // 儲存庫類型，例如："npm", "gradle", "maven"
	PAT       string `yaml:"pat"`       // Personal Access Token
	Base64PAT string `yaml:"base64Pat"` // Base64 編碼的 Personal Access Token
	Username  string `yaml:"username"`  // 用於認證的使用者名稱
	Email     string `yaml:"email"`     // (NPM專用) 使用者電子郵件
	Url       string `yaml:"url"`       // 儲存庫的基礎 URL
	Name      string `yaml:"name"`      // 儲存庫名稱 (例如 Gradle/Maven 的 'musasiyang')
	ID        string `yaml:"id"`        // 儲存庫 ID (例如 Maven settings.xml 的 'musasiyang')
	// 未來可根據需要在此處添加其他通用配置，例如代理設定等。
}

// LoadConfig 從指定路徑載入 YAML 配置檔，並將其內容映射到 RepositoryConnection 結構。
// 如果載入或解析失敗，會列印錯誤訊息並返回原始的 rc 實例。
func (rc *RepositoryConnection) LoadConfig(repoType string, confPath string) *RepositoryConnection {
	fileContent, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Printf("讀取配置檔 %s 失敗: %v\n", confPath, err)
		return rc
	}

	err = yaml.Unmarshal(fileContent, rc) // 使用 yaml.Unmarshal
	if err != nil {
		fmt.Printf("解析配置檔 %s 為 YAML 失敗: %v\n", confPath, err)
		return rc
	}

	fmt.Printf("成功從 %s 載入配置。\n", confPath)
	return rc
}
