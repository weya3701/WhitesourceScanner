# 將套件發布到 Azure DevOps Artifact Feed

## 功能描述
此功能旨在允許應用程式將套件（如 npm, PyPI, Maven 等）發布到 Azure DevOps 中的 Artifact Feed。

## 準備條件

### 1. Azure DevOps 個人存取權杖 (PAT) 設定
*   **生成 PAT**：在 Azure DevOps 中生成具有 `Packaging > Read & Write` 範圍的個人存取權杖。
*   **安全儲存**：PAT 必須安全地儲存。
    *   **建議方式**：考慮將 PAT 作為 Kubernetes Secret 進行管理（如果應用程式運行在 Kubernetes 環境中）。否則，可透過環境變數注入。
    * pat: 

### 2. 定義支援的套件類型及對應工具
*   **明確套件類型**：確定需要支援的套件類型。
    *   **優先考慮**：根據現有程式碼基礎，可優先支援 npm 和 PyPI。
*   **客戶端工具**：確保執行套件發布的環境（例如 Docker 容器）中安裝了相應的客戶端工具：
    *   **npm 套件**：需要 `npm` CLI 工具。
    *   **PyPI 套件**：需要 `python` 和 `twine` CLI 工具。
    *   **Maven 套件**：如果支援，則需要 `java` 和 `maven` CLI 工具。

### 3. 程式碼配置與結構規劃
*   **新增配置**：在程式碼中新增配置來儲存 Azure DevOps 相關資訊 (URL、專案、Feed 名稱)。可以考慮在 `wss/models.go` 中新增結構，或建立專門的配置服務。
*   **發布邏輯**：規劃不同套件類型的發布邏輯，這通常涉及執行對應的 CLI 命令並處理認證（例如，配置 `.npmrc`、`pip.ini` 或 `settings.xml`）。
*   **錯誤處理**：建立健全的錯誤處理機制，以應對網路、認證或命令執行錯誤。
*   **組織 URL**：明確 Azure DevOps 組織的完整 URL (例如：`https://dev.azure.com/YourOrganization`)。
*   **專案名稱**：明確目標專案的名稱 (例如：`YourProject`)。
*   **Feed 名稱**：明確要發布套件的目標 Feed 名稱 (例如：`YourFeed`)。

** 現有架構
* worker/
    * gradle.go
    * handler.go
    * mvn.go
    * npm.go
    * pypi.go
    * urlget.go
