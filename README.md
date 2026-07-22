# WhitesourceScanner

### Mend Whitesource scanner

* * *

## 環境

* Golang 1.24.0

## Build

        ~ go build -o WhitesourceScanner

## 設定

* 設定 conf.yaml，設定項目：

    * apiKey, userKey, productName, productToken

* 透過 `.env` 設定執行路徑、Mend API URL、暫存目錄與外部工具命令。程式啟動時會驗證目前模式所需的設定，缺少設定時會以非零狀態結束。

* 若 `.env` 不存在，程式會使用內建預設值。若 `config/conf.yaml` 不存在，Mend 憑證會改讀取 `MEND_API_KEY`、`MEND_USER_KEY`、`MEND_PRODUCT_TOKEN` 環境變數；其他項目使用內建預設值。檔案存在時仍優先使用檔案內容。

## 執行

* 將套件依目錄放置./tmp目錄中

* 執行以下命令，project_name, package_name填上./tmp/<掃描套件目錄名稱>

        ~ WhitesourceScanner --mode=cmd --project_name=<project_name> --package_name=<package_name>

* cmd 模式也可以用 `scan_source` 直接指定要掃描的目錄。指定後不使用 `project_name`、`package_name`，也不會在 `package_tmp` 下組合路徑；Mend 專案名稱與報告目錄名稱會使用來源目錄的名稱。

        ~ WhitesourceScanner --mode=cmd --scan_source=/path/to/source

* 執行完成在./report/<掃描套件目錄名稱>中可以找到risk.pdf檔案

* 執行Docker image tar檔案掃描

        ~ WhitesourceScan --mode=image --project_name=<project_name> --package_name=<package_name> --tar_file=<tar_file_path>

* 執行套件定義檔下載套件並掃描

        ~ WhitesoruceScan --mode=reqfile --project_name=<project_name> --package_name=<package_name> --package_type=<package_type: gradle> --requirements_file=<requrirement_file>

## 開發驗證

        ~ gofmt -w .
        ~ go vet ./...
        ~ go test -race ./...
        ~ go build ./...
