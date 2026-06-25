package repositoryclient

import "fmt"

// Client 是用於向儲存庫發佈 Artifacts 的介面。
type Client interface {
	// Publish 將位於 packageDirPath 的套件發佈到已配置的儲存庫。
	// conn 包含儲存庫連線資訊，packageDirPath 是要發佈套件的目錄路徑。
	Publish(conn *RepositoryConnection, packageDirPath string) error
}

// NewClient 根據給定的儲存庫類型創建一個新的儲存庫客戶端實例。
func NewClient(rc RepositoryConnection) (Client, error) {

	switch rc.Type {
	case "gradle":

		return nil, fmt.Errorf("NPM ADO 客戶端尚未實作")
	}

	// 保留
	// switch repoType {
	// case rc.Type:
	// // case RepositoryTypeNPM:
	// 	// TODO: 在 repositoryclient/ado/ado.go 中實作 NewNpmAdoClient
	// 	// return ado.NewNpmAdoClient()
	// 	return nil, fmt.Errorf("NPM ADO 客戶端尚未實作")
	// case RepositoryTypeGradle:
	// 	// TODO: 在 repositoryclient/ado/ado.go 中實作 NewGradleAdoClient
	// 	// return ado.NewGradleAdoClient()
	// 	return nil, fmt.Errorf("Gradle ADO 客戶端尚未實作")
	// case RepositoryTypeMaven:
	// 	// TODO: 在 repositoryclient/ado/ado.go 中實作 NewMavenAdoClient
	// 	// return ado.NewMavenAdoClient()
	// 	return nil, fmt.Errorf("Maven ADO 客戶端尚未實作")
	// // case RepositoryTypeNexus: // 未來可擴展
	// //     // return nexus.NewClient()
	// //     return nil, fmt.Errorf("Nexus 客戶端尚未實作")
	// default:
	// 	return nil, fmt.Errorf("不支援的儲存庫類型: %s", repoType)
	// }

	return nil, fmt.Errorf("NPM ADO 客戶端尚未實作")
}
