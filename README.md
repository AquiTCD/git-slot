# Git Slot

**Git Worktree を「固定スロット」として管理する CLI ツール**

Git Slot は、`git worktree` を TOML 設定で定義された固定名のスロットに割り当てて管理する。バイナリ名は `git-slot` で、`PATH` に配置することで `git slot` として Git のサブコマンドとして動作する。

ブランチ中心の運用から **スロット中心** の運用へシフトし、IDE のパス固定やビルドキャッシュの安定化を実現する。

## 解決する課題

- ブランチ名ベースの worktree はパスが毎回変わり、IDE 設定やビルドキャッシュが壊れる
- ブランチ切り替えのたびに環境の再構築が必要
- 命名規則なしに worktree を作ると管理が煩雑になる

## 特徴

- **固定ワークスペース** — スロットは設定で定義された固定名を持ち、パスが安定する
- **設定ベース** — スロットの名前・数はすべて TOML で定義。プリセットなし
- **gwq 互換** — gwq のディレクトリ規約（`~/worktrees/{host}/{owner}/{repo}/`）と共存
- **階層型設定** — プロジェクト固有設定がグローバル設定をオーバーライド
- **安全ガード** — ブランチ重複検出、dirty 状態の保護
- **Git サブコマンド** — `git slot` として自然に使える

## インストール

`git-slot` バイナリを `PATH` の通った場所に配置する。Git が自動的に `git slot` コマンドとして認識する。

```bash
# ショートカットを設定する場合（オプション）
alias gs='git slot'
```

## クイックスタート

```bash
# 設定ファイルを生成
git slot --init

# git-slot.toml を編集してスロットを定義
# （スロット名は自由に設定）

# スロットにブランチを装填
git slot main-work feature/nice-ui

# 新規ブランチを作成して装填
git slot hotfix -c hotfix/urgent-bug

# スロット一覧を確認
git slot --list

# スロットのパスを取得して移動
cd $(git slot main-work)

# スロットを解除
git slot -d main-work

# スロット間でブランチを入れ替え
git slot --swap main-work hotfix
```

## 設定

設定は以下の優先順位でマージされる（後勝ち）:

1. **Global Config** — `~/.config/git-slot/config.toml`
2. **Project Config** — `<project-root>/git-slot.toml`

### 設定例 (`git-slot.toml`)

```toml
# gwq_basedir = "~/worktrees"  # gwq の basedir と同じ値（デフォルト: ~/worktrees）

[[slots]]
name = "main-work"

[[slots]]
name = "hotfix"

[[slots]]
name = "experiment"
```

## ディレクトリ構造

gwq のディレクトリ規約（`~/worktrees/{host}/{owner}/{repo}/`）に準拠し、`slots/` サブディレクトリで共存する。

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
| `git slot <slot> <branch>` | スロットにブランチを装填 |
| `git slot <slot> -c <branch>` | 新規ブランチを作成して装填（`-b` も可） |
| `git slot <slot>` | スロットのパスを出力 |
| `git slot -l, --list` | スロット一覧を表示 |
| `git slot -d, --clear <slot>` | スロットを解除 |
| `git slot -s, --swap <A> <B>` | スロット間のブランチを入れ替え |
| `git slot --status [slot]` | スロットの詳細状態を表示 |
| `git slot --init` | 設定ファイルのテンプレートを生成 |
| `git slot --version` | バージョン情報を表示 |

## 技術スタック

| カテゴリ | 技術 |
|----------|------|
| 言語 | Go |
| CLI | [Cobra](https://github.com/spf13/cobra) |
| 設定 | [pelletier/go-toml](https://github.com/pelletier/go-toml) |

## ドキュメント

詳細な仕様は `docs/specs/` を参照:

- [プロダクト概要](docs/specs/overview.md)
- [コアスロット管理](docs/specs/core-slot-management.md)
- [設定システム](docs/specs/config-system.md)
- [CLI インターフェース](docs/specs/cli-interface.md)
- [外部統合](docs/specs/integration.md)

## ライセンス

MIT
