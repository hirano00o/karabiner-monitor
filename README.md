# Karabiner Monitor

macOS用のkarabiner_grabberプロセスのメモリ監視・自動再起動ツール

## 概要

Karabiner-Elementsの`karabiner_grabber`プロセスは、画面スリープなどを経るとメモリ使用率が上昇し、ショートカットが効かなくなることがあります。このツールは、メモリ使用率を監視し、閾値を超えた場合に自動的にプロセスを再起動することで、この問題を解決します。

## 主な機能

- **メモリ監視**: 設定可能な間隔で`karabiner_grabber`プロセスのメモリ使用量をチェック
- **自動再起動**: メモリ閾値を超えた場合に自動的にプロセスをkill（Karabinerが自動的に再起動）
- **キー入力検知**: プロセスをkillする前に、指定秒数間キー入力がないことを確認
- **macOS通知**: プロセス再起動時に通知センターで通知
- **ログローテーション**: 設定可能なサイズと保持日数でログを自動ローテーション
- **LaunchDaemon**: システム起動時に自動起動（rootプロセス監視のためroot権限で実行）

## 必要要件

- macOS 12以降
- Go 1.21以降（ビルド時）
- Karabiner-Elements
- アクセシビリティ権限

## インストール

### 1. リポジトリのクローン

```bash
git clone https://github.com/hirano00o/karabiner-monitor.git
cd karabiner-monitor
```

### 2. ビルドとインストール

```bash
make install
```

このコマンドは以下を実行します:
- バイナリのビルド
- `/usr/local/bin`へのインストール
- 設定ファイルの生成（`/Library/Application Support/karabiner-monitor/config.json`）
- LaunchDaemon の登録と起動（root権限で実行）

### 3. アクセシビリティ権限の設定

**重要**: キー入力監視にはアクセシビリティ権限が必要です。

1. システム環境設定 > セキュリティとプライバシー > プライバシー を開く
2. 左側のリストから「アクセシビリティ」を選択
3. 鍵アイコンをクリックして変更を許可
4. `/usr/local/bin/karabiner-monitor` を追加
5. チェックボックスを有効化
6. サービスを再起動: `sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist && sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist`

**注意**: このサービスはLaunchDaemonとしてroot権限で実行されます。これは`karabiner_grabber`がrootプロセスとして動作しているためです。

## 設定

設定ファイル: `/Library/Application Support/karabiner-monitor/config.json`

```json
{
  "process_name": "karabiner_grabber",
  "memory_threshold_mb": 50,
  "check_interval_seconds": 60,
  "idle_wait_seconds": 10,
  "log_max_size_mb": 10,
  "log_max_age_days": 7
}
```

### 設定項目

- **process_name**: 監視するプロセス名（デフォルト: `karabiner_grabber`）
- **memory_threshold_mb**: メモリ使用量の閾値（MB）（デフォルト: 50）
- **check_interval_seconds**: メモリチェック間隔（秒）（デフォルト: 60）
- **idle_wait_seconds**: killする前のキー入力待機時間（秒）（デフォルト: 10）
- **log_max_size_mb**: ログファイルの最大サイズ（MB）（デフォルト: 10）
- **log_max_age_days**: ログファイルの保持日数（デフォルト: 7）

## 使用方法

### サービスの状態確認

```bash
sudo launchctl list | grep karabiner.monitor
```

### ログの確認

```bash
# アプリケーションログ
sudo tail -f /var/log/karabiner-monitor/monitor.log

# 標準出力
sudo tail -f /var/log/karabiner-monitor.stdout

# 標準エラー出力
sudo tail -f /var/log/karabiner-monitor.stderr
```

### サービスの停止

```bash
sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
```

### サービスの開始

```bash
sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
```

### ローカル実行（デバッグ用）

```bash
# rootで実行する必要があります
sudo /usr/local/bin/karabiner-monitor
```

## アンインストール

```bash
make uninstall
```

設定ファイルとログを完全に削除する場合:

```bash
sudo rm -rf "/Library/Application Support/karabiner-monitor"
sudo rm -rf /var/log/karabiner-monitor
sudo rm -f /var/log/karabiner-monitor.stdout
sudo rm -f /var/log/karabiner-monitor.stderr
```

## 開発

### ビルド

```bash
make build
```

### テスト実行

```bash
make test
```

### Lint実行

```bash
make lint
```

### コードフォーマット

```bash
make fmt
```

### すべてのチェック実行

```bash
make check
```

## アーキテクチャ

プロジェクト構造:

```
karabiner-monitor/
├── cmd/karabiner-monitor/   # メインアプリケーション
├── internal/
│   ├── config/               # 設定管理
│   ├── logger/               # ログシステム
│   ├── monitor/              # プロセス監視
│   ├── keyboard/             # キー入力監視（CGo）
│   └── notifier/             # macOS通知
├── scripts/                  # インストール/アンインストールスクリプト
├── configs/                  # LaunchAgent plist
└── Makefile                  # ビルド・デプロイ自動化
```

### 主要コンポーネント

1. **Config**: JSON設定ファイルの読み込みと検証
2. **Logger**: lumberjackを使用したログローテーション
3. **Monitor**: gopsutil/v4を使用したプロセス監視
4. **Keyboard**: CGEventTap APIを使用したキー入力監視
5. **Notifier**: osascriptを使用したmacOS通知

## ライセンス

MIT License

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## トラブルシューティング

### アクセシビリティ権限がない

ログに「accessibility permission not granted」と表示される場合は、上記のアクセシビリティ権限の設定を確認してください。

### プロセスが見つからない（process not found）

以下の原因が考えられます:

1. **karabiner_grabberが実行されていない**: プロセスが起動すると自動的に監視が開始されます
2. **LaunchDaemonが正しく起動していない**: サービスの状態を確認してください
   ```bash
   sudo launchctl list | grep karabiner.monitor
   ```
3. **アクセシビリティ権限がない**: アクセシビリティ権限を設定後、サービスを再起動してください
   ```bash
   sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
   sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
   ```

### 通知が表示されない

macOSの通知設定で、ターミナルまたは`karabiner-monitor`の通知が許可されているか確認してください。

## 参考

- [Karabiner-Elements](https://karabiner-elements.pqrs.org/)
- [gopsutil](https://github.com/shirou/gopsutil)
- [lumberjack](https://github.com/natefinch/lumberjack)
