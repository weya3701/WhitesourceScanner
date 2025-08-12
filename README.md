# WhitesourceScanner

### Mend Whitesource scanner

* * *

## 環境

* Golang 1.24.0

## Build

        ~ go build -o WhitesourceScanner

## 設定

* 設定conf.yaml，設定項目：

    * apiKey, uerKye, productName, productToken

## 執行

* 將套件依目錄放置./tmp目錄中

* 執行以下命令，project_name, package_name填上./tmp/<掃描套件目錄名稱>

        ~ WhitesourceScanner --mode=cmd --project_name=<project_name> --package_name=<package_name>

* 執行完成在./report/<掃描套件目錄名稱>中可以找到risk.pdf檔案

* 執行Docker image tar檔案掃描

        ~ WhitesourceScan --mode=image --project_name=<project_name> --package_name=<package_name> --tar_file=<tar_file_path>
