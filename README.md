# Karabiner Monitor

macOS用のkarabiner_grabberプロセスのメモリ監視・自動再起動ツール

## 概要

Karabiner-Elementsの`karabiner_grabber`プロセスは、画面スリープなどを経るとメモリ使用率が上昇し、ショートカットが効かなくなることがあります。このツールは、メモリ使用率を監視し、閾値を超えた場合に自動的にプロセスを再起動することで、この問題を解決します。

## 主な機能

- **メモリ監視**: 設定可能な間隔で`karabiner_grabber`プロセスのメモリ使用量をチェック
- **自動再起動**: メモリ閾値を超えた場合に自動的にプロセスをkill（Karabinerが自動的に再起動）
- **待機時間**: プロセスをkillする前に、設定された秒数だけ待機（デフォルト10秒）
- **macOS通知**: terminal-notifierを使用した通知（推奨）、osascriptへのフォールバック対応
- **詳細なログ**: すべての監視・再起動アクションを記録
- **ログローテーション**: 設定可能なサイズと保持日数でログを自動ローテーション
- **LaunchDaemon**: システム起動時に自動起動（rootプロセス監視のためroot権限で実行）
- **プロセス検索**: プロセス名とコマンドライン引数の両方で検索可能

## 必要要件

- macOS 12以降
- Go 1.21以降（ビルド時）
- Karabiner-Elements
- terminal-notifier（通知機能を使用する場合、推奨）

**注意**: このツールはroot権限で実行されるため、アクセシビリティ権限は不要です。

## インストール

### 1. リポジトリのクローン

```bash
git clone https://github.com/hirano00o/karabiner-monitor.git
cd karabiner-monitor
```

### 2. terminal-notifierのインストール（オプション、推奨）

通知機能を使用する場合は、terminal-notifierをインストールしてください:

```bash
brew install terminal-notifier
```

terminal-notifierはLaunchDaemonからの通知に対応しており、rootプロセスからでもユーザーに通知を送ることができます。

### 3. ビルドとインストール

```bash
make install
```

このコマンドは以下を実行します:
- バイナリのビルド
- `/usr/local/bin`へのインストール
- 設定ファイルの生成（`/Library/Application Support/karabiner-monitor/config.json`）
- LaunchDaemon の登録と起動（root権限で実行）

### 4. 動作確認

インストールが完了すると、LaunchDaemonとしてサービスが起動します。

**重要な仕様**:
- このサービスはLaunchDaemonとしてroot権限で実行されます
- `karabiner_grabber`がrootプロセスとして動作しているため、root権限が必要です
- rootプロセスとして実行されるため、アクセシビリティ権限の設定は不要です
- キーボードアイドル検知は、root環境では実際の監視を行わず、設定された秒数（デフォルト10秒）だけ待機します
- 通知機能は`terminal-notifier`を使用（インストールされている場合）、フォールバックとして`osascript`を使用

ログでサービスの動作を確認できます:

```bash
sudo tail -f /var/log/karabiner-monitor/monitor.log
```

正常に動作している場合、以下のようなログが表示されます:

```
{"level":"INFO","msg":"karabiner-monitor starting","config":"/Library/Application Support/karabiner-monitor/config.json"}
{"level":"INFO","msg":"running as root, skipping accessibility permission check"}
{"level":"INFO","msg":"starting monitoring loop","process":"karabiner_grabber","threshold_mb":50}
{"level":"INFO","msg":"process found","pid":99849,"name":"karabiner_grabber"}
{"level":"INFO","msg":"memory usage","pid":99849,"memory_mb":97.1,"threshold_mb":50}
```

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
- **idle_wait_seconds**: killする前の待機時間（秒）（デフォルト: 10）
  - 注: rootプロセスとして実行されるため、実際のキーボード監視は行わず、この秒数だけ待機します
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
3. **Monitor**: gopsutil/v4を使用したプロセス監視とmacOS固有のメモリ計測
   - プロセス名での完全一致検索
   - コマンドライン引数での部分一致検索
   - macOS: `phys_footprint`による正確なメモリ計測（`top`コマンドのMEM列と同じ）
   - その他のOS: RSS（Resident Set Size）を使用
4. **Keyboard**: 待機時間の管理
   - rootプロセスではアクセシビリティAPIを使用せず、単純な時間待機
5. **Notifier**: terminal-notifierを使用したmacOS通知（osascriptへのフォールバック対応）

### 技術的な実装詳細

#### rootプロセスとしての実行

`karabiner_grabber`はrootプロセスとして実行されており、通常のユーザープロセスからはアクセスできません。gopsutilライブラリでrootプロセスの情報を取得しようとすると"invalid argument"エラーが発生します。

この問題を解決するため、karabiner-monitorはLaunchDaemonとしてroot権限で実行されます:

1. **プロセス監視**: rootとして実行されるため、rootプロセス(karabiner_grabber)の情報を取得可能
2. **アクセシビリティ権限**: rootプロセスはアクセシビリティ権限チェックをスキップ
3. **待機時間**: キーボード監視の代わりに、設定された秒数だけ待機してからkill実行

#### プロセス検索

FindProcess関数は2段階で検索を行います:

1. **完全一致検索**: プロセス名が完全に一致するかチェック
2. **コマンドライン検索**: コマンドライン引数に指定文字列が含まれるかチェック

これにより、`/Library/Application Support/org.pqrs/Karabiner-Elements/bin/karabiner_grabber`のようなフルパスで実行されているプロセスも検出できます。

#### 通知システム

LaunchDaemonから通知を送信するには特別な対応が必要です:

1. **terminal-notifier優先**: LaunchDaemonからでも動作する`terminal-notifier`を最初に試行
2. **osascriptへのフォールバック**: `terminal-notifier`が利用できない場合は`osascript`を使用
3. **エラーハンドリング**: 両方が失敗した場合のみエラーを返す

`terminal-notifier`は`brew install terminal-notifier`でインストール可能で、LaunchDaemonからの通知に最適です。

#### メモリ計測

macOSとその他のOSで異なるメモリ計測方法を使用します:

##### macOS (darwin)

macOSでは`phys_footprint`を使用してメモリ使用量を計測します。これは`top`コマンドのMEM列に表示される値と同じで、プロセスが実際に使用しているメモリの正確な表現です。

**phys_footprintとは**:
- macOSカーネルが計算するプロセスの物理メモリフットプリント
- RSS（Resident Set Size）よりも正確な実メモリ使用量
- カーネルの計算式: `(internal - alternate_accounting) + (internal_compressed - alternate_accounting_compressed) + iokit_mapped + purgeable_nonvolatile + purgeable_nonvolatile_compressed + page_table`

**実装方法**:
- `vmmap --summary <pid>`コマンドを使用してPhysical footprintを取得
- rootプロセスとして実行される場合は`vmmap`を直接実行
- 非rootの場合は`sudo -n vmmap`を使用（sudoersでNOPASSWD設定が必要）
- 正規表現でvmmap出力を解析: `Physical footprint:\s+([0-9.]+)([KMGT])?`
- K/M/G/T単位を自動的にMBに変換

**RSSとの比較**:
- RSS: 物理メモリに常駐しているページのサイズ（共有メモリを含む）
- phys_footprint: プロセスが実際に使用している物理メモリ（共有メモリの正確な割り当てを考慮）
- 例: `karabiner_grabber`の場合、RSSは約13MBだが、phys_footprintは約797MB

**制限事項**:
- `vmmap`コマンドの実行には約2秒かかる場合があります
- rootプロセスに対しては、実行側もroot権限が必要です

##### その他のOS (Linux, Windows)

非macOSプラットフォームでは、gopsutilライブラリを使用してRSS（Resident Set Size）を取得します:
- RSS: プロセスが物理メモリに保持しているメモリのサイズ
- バイトからメガバイトに変換して返します

## ライセンス

MIT License

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## トラブルシューティング

### プロセスが見つからない（process not found）

以下の原因が考えられます:

1. **karabiner_grabberが実行されていない**: プロセスが起動すると自動的に監視が開始されます
   ```bash
   ps aux | grep karabiner_grabber | grep -v grep
   ```

2. **LaunchDaemonが正しく起動していない**: サービスの状態を確認してください
   ```bash
   sudo launchctl list | grep karabiner.monitor
   ```

3. **サービスが起動に失敗している**: エラーログを確認してください
   ```bash
   sudo tail -50 /var/log/karabiner-monitor.stderr
   ```

### サービスが頻繁に再起動している

`/var/log/karabiner-monitor.stderr`に繰り返しエラーが記録されている場合:

1. ログを確認してエラー内容を特定
2. 設定ファイルが正しいか確認: `/Library/Application Support/karabiner-monitor/config.json`
3. サービスを停止して問題を解決してから再起動:
   ```bash
   sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
   # 問題を修正
   sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
   ```

### メモリ閾値を超えてもkillされない

1. ログでメモリ使用量を確認:
   ```bash
   sudo tail -f /var/log/karabiner-monitor/monitor.log
   ```

2. `idle_wait_seconds`の待機時間後にkillされます（デフォルト10秒）

3. killに成功すると以下のようなログが表示されます:
   ```
   {"level":"INFO","msg":"killing process","pid":99849,"memory_mb":97.1}
   {"level":"INFO","msg":"process killed successfully","pid":99849}
   ```

### 通知が表示されない

**推奨**: terminal-notifierをインストールすることで、LaunchDaemonからでも通知が表示されます。

```bash
brew install terminal-notifier
```

terminal-notifierがインストールされていない場合、osascriptにフォールバックしますが、rootプロセスからの通知は制限される場合があります。

**通知が表示されない場合の確認方法**:

1. **terminal-notifierがインストールされているか確認**:
   ```bash
   which terminal-notifier
   ```

2. **ログで通知送信のエラーを確認**:
   ```bash
   sudo tail -50 /var/log/karabiner-monitor.stderr | grep notification
   ```

3. **ログファイルで再起動を確認**（通知の代替）:
   ```bash
   sudo tail -f /var/log/karabiner-monitor/monitor.log | grep "killed"
   ```

4. **karabiner_grabberのPID変化を確認**:
   ```bash
   # 再起動前のPIDを記録
   ps aux | grep karabiner_grabber | grep -v grep

   # しばらく待ってから再度確認
   # PIDが変わっていれば再起動された
   ```

**注意**: terminal-notifierをインストール後は、サービスの再起動が必要です:
```bash
sudo launchctl unload /Library/LaunchDaemons/com.karabiner.monitor.plist
sudo launchctl load /Library/LaunchDaemons/com.karabiner.monitor.plist
```

## 参考

- [Karabiner-Elements](https://karabiner-elements.pqrs.org/)
- [gopsutil](https://github.com/shirou/gopsutil)
- [lumberjack](https://github.com/natefinch/lumberjack)
