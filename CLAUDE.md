1. 目的と実現したいこと

メインの目的: 外出先（スーパーなど）からスマホのWebアプリで調味料置き場の最新写真をチェックし、買い忘れや二重買いを防ぐ 。


撮影アプローチ: 定期撮影ではなく、スマホから「撮影ボタン」を押したタイミングでシャッターを切る「オンデマンド方式」を採用し、通信量やクラウド容量を節約する 。

2. ハードウェア構成

エッジデバイス: Raspberry Pi（お手持ちの機材を活用） 。


カメラ: Logicool C270nd 。

約1,980円の1年保証高コスパモデル 。

720pの解像度と固定フォーカス（40cm以上で全体にピントが合う）により、調味料の残量を十分に判別可能 。

Linux標準のUVCドライバ対応のため、USBポートに挿すだけで認識される 。


ストレージ: 高耐久（High Endurance）タイプのMicroSDカード（手配済み） 。


電源: ラズパイのモデルに適合した専用の安定化電源（負荷上昇時の電圧降下・再起動を防ぐため必須 / 手配済み） 。

設置・固定方法:

コンロ周辺の熱（上限50℃程度）や油煙によるショート・発火リスクを避けるため、対面カウンターや冷蔵庫の上など、離れた安全な場所から見下ろすように設置する 。

100円ショップのスマホスタンドやテープ、結束バンド等を使用してカメラの角度を固定する 。


被写体（調味料）の工夫: 残量が見えやすいように透明・半透明の容器に統一し、奥のボトルが隠れないようにひな壇状などに並べる 。

3. ソフトウェア・クラウドアーキテクチャ
システム全体のデータの流れは 「Webアプリ ➡️ API Gateway ➡️ Lambda ➡️ IoT Core ➡️ ラズパイ ➡️ S3 ➡️ Webアプリ（署名付きURL）」 となります 。

エッジ側（ラズパイ）の実装

実装言語: Go言語（メモリ消費が少なく、シングルバイナリでデプロイ可能） 。

処理フロー:

起動時からAWS IoT CoreとMQTTで接続を維持し、特定のトピックをサブスクライブして待機する 。

トリガーを受信したら、fswebcamコマンドを実行して静止画を撮影する 。

撮影した画像をAWS SDKを使用してS3バケットへ直接アップロードする 。


ネットワーク設定: OS書き込み時にRaspberry Pi ImagerでWi-FiとSSHを事前設定し、ヘッドレス（ディスプレイなし）で起動させる 。外部からのSSH（ポート22）は公開せず、アウトバウンド通信のみの構成にする 。

バックエンド（AWS側）の実装

実装言語 (Lambda): Go言語（コールドスタートが爆速でラグが少なく、エッジ側とシステム全体の言語を統一できる） 。


メッセージング: AWS IoT Core（個人用途なら無料枠に収まり、完全従量課金のため非常に安価） 。


ストレージ: Amazon S3（非公開バケットとして構築し、インターネットから直接アクセスできないようブロックする） 。


インフラ管理: Terraformを利用して、S3、IoT Core、IAMポリシーなどのAWSリソースをコードベースでプロビジョニングする 。


運用コスト: 12ヶ月の無料枠終了後も、システム全体の維持費は月に15円〜20円程度に収まるサーバーレス構成 。

セキュリティと認証

APIの保護: 誰でもシャッターを切れないように、AWS Cognitoによるユーザー認証、またはLambdaオーソライザー（独自認証・Basic認証など）をAPI Gatewayに組み込む 。


画像の保護: 認証を通過したリクエストに対してのみ、Lambdaが数分間だけ有効な「S3の署名付きURL」を発行し、スマホアプリはそれを使って安全に画像を表示する 。


権限の最小化: ラズパイにはIoT Coreのデバイス証明書を利用し、「対象S3へのPutObject」と「特定MQTTトピックの送受信」のみを許可する最小権限のIAMポリシーを付与する 。

4. 将来の拡張構想（ロードマップ）

AI画像認識（Amazon Rekognition）の導入: S3にアップロードされた画像を自動で解析し、「マヨネーズ残量何％」といったテキストデータ（JSON）化してアプリのUIに警告表示などを出す（最初の12ヶ月は無料枠で月に数千枚まで実質タダ） 。


マルチカメラ化: セルフパワー型のUSBハブを追加し、複数のUSBカメラをラズパイに接続して、アプリ側から「引きの絵」と「アップの絵」の撮影を切り替える 。

---

## 5. リポジトリ構成（モノレポ）

```
home_device/
├── edge/        ラズパイ用 Go エージェント（MQTT subscribe → 撮影 → S3 アップロード）
├── lambda/      Go 製 Lambda（trigger: 撮影トリガ publish / image: 署名付きURL発行）
├── terraform/   AWS リソース（S3 / IoT Core / Cognito / API Gateway / Lambda / IAM）
└── frontend/    React + Vite (TypeScript) のスマホ向け SPA
```

`edge` と `lambda` は別デプロイ単位のため独立 Go モジュールとし、ルートの `go.work` で束ねる。

## 6. データフロー

```
frontend(React) →[JWT]→ API Gateway(HTTP API + Cognito JWT authorizer)
  ├─ POST /capture → lambda(trigger) → IoT Core publish: home_device/{deviceId}/capture
  │                                       → edge が subscribe → fswebcam 撮影 → S3 PutObject
  └─ GET  /image   → lambda(image)   → S3 presigned GET URL を発行 → frontend が表示
S3: 非公開バケット（Public Access Block 全有効）。キー例 images/{deviceId}/latest.jpg
```

## 7. 開発コマンド

| 対象 | コマンド |
| --- | --- |
| edge ビルド | `cd edge && go build ./...` |
| edge ローカル実行（実機なし） | `cd edge && CAMERA=mock go run ./cmd/agent` |
| lambda ビルド | `cd lambda && go build ./...` |
| lambda デプロイ用 zip | `cd lambda && make build` |
| terraform 検証 | `cd terraform && terraform init -backend=false && terraform validate` |
| terraform 反映 | `cd terraform && terraform init && terraform apply` |
| frontend 開発 | `cd frontend && npm install && npm run dev` |
| frontend ビルド | `cd frontend && npm run build` |

> 現状の準備状況: AWS アカウント / ラズパイ実機 / カメラ（Logicool C270nd）の手配・接続まで完了済み（`fswebcam` での撮影動作も確認済み）。独自ドメインのみ未手配。AWS 側のリソース構築（terraform apply・Lambda デプロイ・IoT デバイス証明書発行）はこれから。開発マシン上では `edge` を `CAMERA=mock` で起動して擬似動作を確認することも引き続き可能。