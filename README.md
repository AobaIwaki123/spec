# db-quiz

IPA高度情報処理技術者「データベーススペシャリスト試験（DB）」の午前IIおよび午後試験対策を支援するGo言語製CLIツールです。

## 主な機能

1. **午前II 択一クイズ & 弱点克服**
   - ランダム出題（デフォルト5問）、分野別特訓、出題数指定
   - 直近で間違えた問題のみを集中的に反復演習する弱点復習モード（`-r` / `review`）
2. **学習進捗 & 午前II予想得点推論モニタリング**
   - 累計回答数、正解率、分野別習熟度判定
   - 本番出題比率と直近回答を加重平均した「午前II 予想得点（100点満点換算）」および合格可能性推論
3. **IPA公式 過去問PDF自動取得（非同期・バックグラウンド対応）**
   - 令和2年度〜令和6年度の午前II、午後I、午後II問題冊子・解答例・採点講評PDFのダウンロード
   - バックグラウンド実行（`-d` / `--background`）とステータスモニタリング（`--status`）
4. **チャットエージェント連携インターフェース**
   - チャット上での問題出題用JSON出力（`question --json`）
   - 回答結果を1行で自動補完・永続化する記録コマンド（`record --id <id> --choice <0-3>`）

---

## インストール方法

PATHが通っている `~/.local/bin` 配下にバイナリをインストールします。

```bash
# クローンまたはディレクトリへ移動
cd ~/db-quiz

# インストールスクリプトを実行
./scripts/install.sh

# または make install
make install
```

インストール後、ターミナル上の任意のディレクトリから `db-quiz` だけで起動可能です。

---

## コマンド一覧

### 1. クイズの実行

```bash
# 通常出題（ランダム5問）
db-quiz

# 出題数を指定（例: 10問）
db-quiz -n 10

# 分野を絞って特訓（例: 正規化理論）
db-quiz -c "正規化理論"

# 直近の間違い問題を集中的に復習（弱点克服モード）
db-quiz -r
# または
db-quiz review

# 分野一覧と問題数の確認
db-quiz categories
```

### 2. 学習統計・予想得点レポート

```bash
db-quiz stats
# または
db-quiz --stats
```

### 3. 過去問PDFの取得（fetch）

```bash
# 全年度（令和2年〜令和6年）をバックグラウンドで非同期取得
db-quiz fetch --all -d

# 進行状況と取得済みファイル一覧を確認
db-quiz fetch --status

# 特定年度のみ同期取得（例: 令和6年）
db-quiz fetch --year 2024

# 午前IIのみに絞って取得
db-quiz fetch --type am2
```

### 4. チャットエージェント連携（record / question）

```bash
# 問題データをJSONで取得（チャット出題用）
db-quiz question --json
db-quiz question --category "正規化理論" --json
db-quiz question --review-wrong --json

# 回答結果を1行で記録（問題DBからカテゴリ・正解・正誤を自動解決）
db-quiz record --id db-011 --choice 2
```

---

## ディレクトリ構成

```
db-quiz/
├── Makefile                     # build, install, test, verify
├── README.md                    # 本ドキュメント
├── go.mod                       # Go 1.26.4
├── cmd/
│   └── db-quiz/
│       └── main.go              # CLIエントリポイント・サブコマンド処理
├── internal/
│   ├── quiz/                    # 出題・採点・統計推論・履歴保存
│   │   ├── model.go
│   │   ├── store.go
│   │   ├── stats.go
│   │   ├── runner.go
│   │   ├── recorder.go
│   │   ├── stats_test.go
│   │   └── store_test.go
│   └── fetcher/                 # IPA過去問アーカイブ取得・非同期管理
│       ├── ipa.go
│       ├── manager.go
│       └── fetcher_test.go
├── data/
│   ├── questions.json           # 午前II問題データ
│   ├── quiz_history.jsonl       # 回答履歴ログ
│   ├── fetch_status.json        # 過去問取得タスク状況
│   └── exams/                   # ダウンロード済みPDF（年度別）
└── scripts/
    ├── install.sh               # ~/.local/bin/ へのインストール
    └── verify-all.sh            # 静的解析・テスト・ビルド検証
```
