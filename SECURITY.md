# セキュリティと実行方法

Go 1.27.2 以上を使用してください。標準ライブラリのセキュリティ修正を
取り込むため、Go 自体も定期的に更新してください。

## 認証情報

`cmd/aws` は AWS SDK の標準認証情報チェーンに対応しています。
`AWS_ACCESS_KEY_ID` と `AWS_SECRET_ACCESS_KEY`、または AWS の共有
認証情報ファイル・プロファイルを設定してから実行してください。
一時認証情報の場合は `AWS_SESSION_TOKEN` も設定できます。

```sh
go run ./cmd/aws -b BUCKET -k OBJECT_KEY < FILE
```

`-a` と `-s` は互換性のため利用できますが、両方を指定する必要があります。
コマンド引数に秘密鍵を渡すと、シェル履歴やプロセス一覧に残る可能性が
あるため、環境変数や認証情報ファイルの利用を推奨します。
実際の認証情報をリポジトリに追加しないでください。

## アップロード

接続先は `https://ds.jp-east.idcfcloud.com` です。
`cmd/goamz` のサンプルは `private` ACL でアップロードします。
今回の変更は既存オブジェクトの ACL やバケットポリシーを変更しません。
バケットポリシーなどによる公開設定は別途管理してください。

`cmd/goamz` が利用する
[mitchellh/goamz](https://github.com/mitchellh/goamz) はアーカイブ済みです。
新しい実装には、保守されている AWS SDK v2 を使用する `cmd/aws` を利用してください。

## 検査

```sh
go mod verify
go test ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
```

GitHub Actions で push・pull request・週次の検査を実施します。
Dependabot は Go モジュールと GitHub Actions の更新 PR を週次で作成します。
検査ツールのバージョンも定期的に見直してください。
