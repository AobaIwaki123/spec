# spec

IPA高度情報処理技術者・スペシャリスト試験 CLI

## インストール

```bash
make install
```

## 使い方

```bash
# クイズ出題（選択肢常時シャッフル・デフォルトDB）
spec
spec nw
spec -n 10
spec -c "正規化理論"
spec -r

# 弱点・予想得点レポート
spec stats
spec nw stats

# 過去問PDFバックグラウンド取得
spec fetch --all -d
spec fetch --status

# チャット・外部連携
spec question --json
spec record --id db-011 --choice 2

# 対応試験一覧
spec exams
```
