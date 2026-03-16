# Git Slot

**Git Worktree を「固定スロット」として管理する CLI ツール**

Git Slot は、`git worktree` を TOML 設定で定義された固定名のスロットに割り当てて管理します。バイナリ名は `git-slot` で、`PATH` に配置することで `git slot` として Git のサブコマンドとして動作します。

ブランチ中心の運用から **スロット中心** の運用へシフトし、IDE のパス固定やビルドキャッシュの安定化を実現します。

## 解決する課題

- ブランチ名ベースの worktree はパスが毎回変わり、IDE 設定やビルドキャッシュが壊れます
- ブランチ切り替えのたびに環境の再構築が必要になります
- 命名規則なしに worktree を作ると管理が煩雑になります

## 特徴

- **固定ワークスペース** — スロットは設定で定義された固定名を持ち、パスが安定します
- **設定ベース** — スロットの名前・数はすべて TOML で定義。プリセットはありません
- **gwq 互換** — gwq のディレクトリ規約（`~/worktrees/{host}/{owner}/{repo}/`）と共存できます
- **階層型設定** — プロジェクト固有設定がグローバル設定をオーバーライドします
- **安全ガード** — ブランチ重複検出、dirty 状態の保護を備えています
- **インタラクティブ TUI** — スロット選択・ブランチ入力を対話的に操作できます（あいまいフィルタ対応）
- **Git サブコマンド** — `git slot` として自然に使えます

## Git Slot の本質

Git の worktree は、ひとつのリポジトリを複数のディレクトリで同時に開ける仕組みです。やっていることは本質的には「同じリポジトリを複数の場所に clone する」のと大きく変わりませんが、`.git` を共有するため容量効率がよく、ブランチ間の操作もスムーズになります。

Git Slot はこの worktree の仕組みを土台にして、以下を提供します:

- **固定スロットによる配置ルール** — ブランチ名に依存しない安定したパス
- **ライフサイクル管理** — 装填・解除・入れ替えの操作と安全ガード
- **Hook による後処理** — スロット操作に連動した自動化スクリプト
- **TUI / CLI ツール** — 日常的な worktree 運用を快適にするインターフェース

「worktree を手動管理する煩雑さ」を設定とツールで吸収し、**スロットという抽象で worktree を扱いやすくする**のが Git Slot の役割です。

## インストール

### Homebrew（推奨）

```bash
brew tap AquiTCD/tap
brew install git-slot
```

### `go install`

```bash
go install github.com/AquiTCD/git-slot/cmd/git-slot@latest
```

`$GOPATH/bin` に PATH が通っていない場合は、シェル設定ファイル（`~/.zshrc` 等）に以下を追加してください:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### ソースからビルド

```bash
git clone https://github.com/AquiTCD/git-slot.git
cd git-slot
make build
```

`./bin/git-slot` が生成されますので、PATH の通った場所にコピーまたはリンクしてください:

```bash
# 例: ~/.local/bin にリンク
ln -sf "$PWD/bin/git-slot" ~/.local/bin/git-slot
```

### 動作確認

`git-slot` が PATH に配置されると、Git が自動的に `git slot` サブコマンドとして認識します。

```bash
git slot --version
```

### cd ラッパー (`gsl`)

`gsl` コマンドを使うと、スロットへの装填・移動を一発で行えます。`git slot` は CLI ツールとして表示・管理を担当し、`gsl` は cd 付きの日常操作用ラッパーです。

```bash
# Bash / Zsh — シェル設定ファイルに追加:
eval "$(git-slot wrapper zsh)"

# Fish:
git-slot wrapper fish | source
```

```bash
# gsl でスロットにブランチを装填 → 自動で cd
gsl set wood feature/nice-ui

# 別のスロットに切り替え（cd も自動）
gsl set fire hotfix/urgent-bug

# 引数なしでインタラクティブモード（TUI でスロット選択）
gsl

# 元のリポジトリに戻る
gsl root
```

## クイックスタート

```bash
# 設定ファイルを生成
git slot init

# git-slot.toml を編集してスロットを定義
# （スロット名は自由に設定できます）

# スロットにブランチを装填
git slot set main-work feature/nice-ui

# 新規ブランチを作成して装填
git slot set hotfix -c hotfix/urgent-bug

# スロット一覧を確認
git slot list

# JSON 形式で出力
git slot list --json

# スロットのパスを取得して移動（gsl なら自動 cd）
gsl set main-work

# スロットを解除
git slot clear main-work

# dirty 状態のスロットを強制解除
git slot clear main-work -f
```

## 設定

設定は以下の優先順位でマージされます（後勝ち）:

1. **Global Config** — `~/.config/git-slot/config.toml`
2. **Project Config** — `<project-root>/git-slot.toml`

### 設定例 (`git-slot.toml`)

```toml
# gwq_basedir = "~/worktrees"  # gwq の basedir と同じ値（デフォルト: ~/worktrees）

[[slots]]
name = "main-work"
icon = "🚀"

[[slots]]
name = "hotfix"
icon = "🔥"

[[slots]]
name = "experiment"

# Optional: hooks
# [hooks]
# post_load = ".git-slot/hooks/post-load.sh"
# post_clear = ".git-slot/hooks/post-clear.sh"

# Optional: TUI settings
# [tui]
# filter = true   # Enable fuzzy filter in interactive mode (default: false)
```

### TUI フィルタ

`[tui] filter = true` を設定すると、インタラクティブモード（`gsl` 引数なし）でリアルタイム絞り込みが有効になります。スロット名やブランチ名の部分一致で素早く選択できます。

## ディレクトリ構造

gwq のディレクトリ規約（`~/worktrees/{host}/{owner}/{repo}/`）に準拠し、`slots/` サブディレクトリで共存します。

```text
~/worktrees/github.com/user/repo/    (gwq の worktree 領域)
├── slots/                            ← Git Slot 専用
│   ├── main-work/
│   ├── hotfix/
│   └── experiment/
├── feature-auth/                     (gwq 通常 worktree)
└── bugfix-login/                     (gwq 通常 worktree)
```

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `git slot` | インタラクティブ TUI を起動します |
| `git slot set <slot> <branch>` | スロットにブランチを装填します |
| `git slot set <slot> -c <branch>` | 新規ブランチを作成して装填します（`-b` も可） |
| `git slot set <slot>` | スロットのパスを出力します |
| `git slot list` | スロット一覧を表示します |
| `git slot clear <slot>` | スロットを解除します |
| `git slot status [slot]` | スロットの詳細状態を表示します |
| `git slot init [-g]` | 設定ファイルのテンプレートを生成します（`-g` でグローバル） |
| `git slot hook [-g]` | フック設定の TUI を起動します（`-g` でグローバル） |
| `git slot root` | リポジトリルートのパスを出力します（`gsl root` で cd） |
| `git slot -v, --version` | バージョン情報を表示します |

### サブコマンド別フラグ

| フラグ | 対象サブコマンド | 説明 |
|--------|------------------|------|
| `--json` | `list`, `status` | 出力を JSON 形式にします |
| `-f, --force` | `set`, `clear`, `init` | dirty 状態の安全チェックをスキップします |
| `-g, --global` | `init`, `hook` | グローバル設定を対象にします |
| `-c, --create` | `set` | 新規ブランチを作成して装填します |
| `-b, --branch` | `set` | `--create` のエイリアスです |

## 技術スタック

| カテゴリ | 技術 |
|----------|------|
| 言語 | Go |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| TUI | [Bubble Tea](https://github.com/charmbracelet/bubbletea) / [Bubbles](https://github.com/charmbracelet/bubbles) |
| 設定 | [pelletier/go-toml](https://github.com/pelletier/go-toml) |
| リリース | [GoReleaser](https://goreleaser.com/) |

## ドキュメント

詳細な仕様は `docs/specs/` を参照してください:

- [プロダクト概要](docs/specs/overview.md)
- [コアスロット管理](docs/specs/core-slot-management.md)
- [設定システム](docs/specs/config-system.md)
- [CLI インターフェース](docs/specs/cli-interface.md)
- [外部統合](docs/specs/integration.md)

## Development

### 必要環境

- Go 1.26+
- [golangci-lint](https://golangci-lint.run/) v2

### ビルド・テスト・Lint

```bash
make build          # ./bin/git-slot にビルド
make test           # go test -race ./...
make lint           # golangci-lint run
make check          # fmt + vet + lint + test を一括実行
```

### その他のターゲット

```bash
make install           # $GOPATH/bin にインストール
make fmt               # gofmt -s -w .
make vet               # go vet ./...
make test-coverage     # カバレッジレポート生成
make release-snapshot  # GoReleaser でローカルスナップショットビルド
make clean             # ビルド成果物の削除
make help              # ターゲット一覧
```

## Acknowledgements

Git Slot は [gwq](https://github.com/d-kuro/gwq) に強くインスパイアされています。gwq の「ホスト/オーナー/リポジトリ」に基づくディレクトリ規約は、worktree の配置を体系化する優れたアプローチであり、Git Slot のディレクトリ構造はこの規約と完全に互換性を持つよう設計しました。gwq が worktree の「配置」を解決したのに対し、Git Slot は固定スロットという抽象で「管理」の側面を補完します。

## ライセンス

MIT
