# CLI インターフェース

## 1. Overview

Git Slot の CLI は `git branch` 型の設計を採用する。メイン操作（スロットへのブランチ装填）は位置引数でシンプルに行い、管理系の操作はフラグで提供する。

バイナリ名は `git-slot` とし、`PATH` に配置することで `git slot` として Git のサブコマンドとして動作する。利便性のため `alias gs='git slot'` の設定も推奨する。

### 設計思想: `git branch` 型パターン

`git branch` が「ブランチ名を引数に取るメイン操作 + フラグで管理操作」という構造であるように、`git slot` も「スロット名を引数に取るメイン操作 + フラグで管理操作」とする。

```bash
# git branch の例（参考）
git branch <name>          # メイン: ブランチ作成
git branch -d <name>       # フラグ: 削除
git branch -m <old> <new>  # フラグ: 改名
git branch --list          # フラグ: 一覧

# git slot の設計
git slot <slot> <branch>   # メイン: スロットにブランチ装填
git slot <slot> -c <new>   # メイン: 新規ブランチ作成して装填
git slot -d <slot>         # フラグ: スロット解除
git slot --swap <A> <B>    # フラグ: 入れ替え
git slot --list            # フラグ: 一覧
```

## 2. PRD (Product Requirements)

### 2.1 User Stories

#### US-CLI-001: 直感的なメイン操作

> 開発者として、`git slot <slot> <branch>` だけでスロットにブランチを装填したい。それにより、日常的なワークフローの摩擦が最小化される。

#### US-CLI-002: スロットパスの取得

> 開発者として、`git slot <slot>` でスロットのパスを取得し、`cd` と組み合わせて移動したい。それにより、スロットへの移動がワンコマンドで行える。

#### US-CLI-003: ヘルプの充実

> 初めて使う開発者として、使い方を簡単に確認したい。それにより、ドキュメントを参照せずに操作を学べる。

#### US-CLI-004: スクリプタブルな出力

> CI/CD パイプラインや他のスクリプトから利用する開発者として、機械可読な出力を得たい。それにより、自動化ワークフローに組み込める。

### 2.2 Acceptance Criteria

| ID | ストーリー | 受け入れ条件 |
|----|-----------|-------------|
| AC-CLI-001 | US-CLI-001 | `git slot <slot> <branch>` でスロットにブランチを装填できる |
| AC-CLI-002 | US-CLI-002 | `git slot <slot>` でスロットの絶対パスが stdout に出力される |
| AC-CLI-003 | US-CLI-003 | `git slot --help` で使い方が表示される |
| AC-CLI-004 | US-CLI-004 | `--json` フラグで JSON 出力が得られる |

### 2.3 Out of Scope

- TUI ダッシュボード（将来のオプション）
- GUI アプリケーション

## 3. TRD (Technical Requirements)

### 3.1 Architecture

#### コマンド体系

```
git slot <slot> <branch>          # スロットに既存ブランチを装填
git slot <slot> -c <branch>       # 新規ブランチを作成して装填（-b も同義）
git slot <slot>                   # スロットのパスを出力

git slot -l, --list               # スロット一覧を表示
git slot -d, --clear <slot>       # スロットを解除
git slot -s, --swap <slot> <slot> # スロット間のブランチを入れ替え
git slot --status [slot]          # スロットの詳細状態を表示
git slot --init [--global]        # 設定ファイルのテンプレートを生成
git slot --version                # バージョン情報を表示
git slot --help                   # ヘルプを表示
```

#### フラグ一覧

| Short | Long | 引数 | 説明 |
|-------|------|------|------|
| `-c` | `--create` | `<name>` | 新規ブランチを作成して装填（`git switch -c` 由来） |
| `-b` | `--branch` | `<name>` | `-c` のエイリアス（`git checkout -b` 由来） |
| `-l` | `--list` | なし | スロット一覧を表示 |
| `-d` | `--clear` | `<slot>` | スロットを解除（`git branch -d` に倣い `-d`） |
| | `--force` | なし | Dirty 状態の確認をスキップ |
| `-s` | `--swap` | `<slot> <slot>` | スロット間のブランチを入れ替え |
| | `--status` | `[slot]` | スロットの詳細状態を表示 |
| | `--init` | なし | 設定ファイルを生成 |
| | `--global` | なし | `--init` と併用。グローバル設定を生成 |
| | `--json` | なし | JSON 形式で出力 |
| | `--version` | なし | バージョン情報を表示 |
| `-h` | `--help` | なし | ヘルプを表示 |

#### 呼び出し方法

バイナリ名が `git-slot` であるため、Git のサブコマンド拡張機構により以下の方法で呼び出せる:

```bash
# Git サブコマンドとして（推奨）
git slot --list
git slot main-work feature/nice-ui

# バイナリ直接実行
git-slot --list
git-slot main-work feature/nice-ui

# エイリアス経由（オプション）
# alias gs='git slot'
gs --list
gs main-work feature/nice-ui
```

### 3.2 Data Model

#### 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 一般エラー（引数不正、操作失敗） |
| 2 | 設定エラー（TOML パース失敗、バリデーションエラー） |
| 3 | Git エラー（リポジトリ外、worktree 操作失敗） |
| 130 | ユーザーによる中断（Ctrl+C） |

### 3.3 Implementation Details

#### 3.3.1 コマンド詳細

##### `git slot <slot> <branch>` — メイン操作: ブランチ装填

```
$ git slot main-work feature/nice-ui
Loading branch 'feature/nice-ui' into slot 'main-work'...
✓ Slot 'main-work' is ready.
  Path: /home/user/src/.../slots/main-work

$ git slot hotfix -c hotfix/urgent-bug
Creating and loading branch 'hotfix/urgent-bug' into slot 'hotfix'...
✓ Slot 'hotfix' is ready.
  Path: /home/user/src/.../slots/hotfix
```

##### `git slot <slot>` — スロットパス出力

ブランチを指定しない場合、スロットの絶対パスを stdout に出力する。`cd` と組み合わせて使う。

```
$ git slot main-work
/home/user/src/.../slots/main-work

# cd と組み合わせ
$ cd $(git slot main-work)
```

スロットが Empty の場合はエラー:

```
$ git slot experiment
Error: slot 'experiment' is empty. Load a branch first: git slot experiment <branch>
```

##### `git slot --list`

全スロットの状態を一覧表示する。

```
$ git slot --list
  main-work   [active]  feature/nice-ui    (a1b2c3d)  *dirty
  hotfix      [active]  hotfix/urgent-bug  (e4f5g6h)
  experiment  [empty]
```

`--json` フラグ付き:

```json
{
  "slots": [
    {
      "name": "main-work",
      "state": "dirty",
      "branch": "feature/nice-ui",
      "head": "a1b2c3d",
      "path": "/home/user/src/.../slots/main-work"
    },
    {
      "name": "hotfix",
      "state": "active",
      "branch": "hotfix/urgent-bug",
      "head": "e4f5g6h",
      "path": "/home/user/src/.../slots/hotfix"
    },
    {
      "name": "experiment",
      "state": "empty",
      "branch": "",
      "head": "",
      "path": "/home/user/src/.../slots/experiment"
    }
  ]
}
```

##### `git slot -d <slot>` / `git slot --clear <slot>`

```
$ git slot -d main-work
Clearing slot 'main-work' (branch: feature/nice-ui)...
✓ Slot 'main-work' is now empty.

$ git slot -d hotfix
Slot 'hotfix' has uncommitted changes.
Continue? [y/N]: n
Cancelled.

$ git slot -d hotfix --force
Force clearing slot 'hotfix'...
✓ Slot 'hotfix' is now empty.
```

##### `git slot --swap <slotA> <slotB>`

```
$ git slot --swap main-work hotfix
Swapping slots...
  main-work: feature/nice-ui → hotfix/urgent-bug
  hotfix:    hotfix/urgent-bug → feature/nice-ui
✓ Swap complete.
```

##### `git slot --status [slot]`

```
$ git slot --status main-work
Slot:    main-work
State:   active (dirty)
Branch:  feature/nice-ui
HEAD:    a1b2c3d (feat: add nice UI component)
Path:    /home/user/src/.../slots/main-work
Changes:
  M  src/components/Button.tsx
  ?? src/components/Modal.tsx
```

引数なしの場合は全スロットの status を表示。

##### `git slot --init`

```
$ git slot --init
Created git-slot.toml with template configuration.

$ git slot --init --global
Created ~/.config/git-slot/config.toml with template configuration.

$ git slot --init
git-slot.toml already exists. Use --force to overwrite.
```

##### `git slot --version`

```
$ git slot --version
git-slot version 0.1.0 (commit: abc1234, built: 2026-03-13)
```

#### 3.3.2 引数解析の優先順位

位置引数とフラグが混在するため、以下の優先順位で解析する:

1. `--version`, `--help` → 即座に出力して終了
2. `--init`, `--list` → 管理操作（位置引数は無視）
3. `--clear <slot>`, `--swap <A> <B>`, `--status [slot]` → フラグ付き管理操作
4. `<slot> <branch>` / `<slot> -c <branch>` → メイン操作: ブランチ装填
5. `<slot>` のみ → パス出力

排他フラグ: `--list`, `--clear`, `--swap`, `--status`, `--init` は互いに排他。同時指定はエラー。
`-c` と `-b` は同義（両方とも新規ブランチ作成）。同時指定はエラー。

#### 3.3.3 出力カラースキーム

| 要素 | 色 | 用途 |
|------|-----|------|
| スロット名 | Bold | 識別性 |
| [active] | Green | 正常状態 |
| [empty] | Gray/Dim | 未使用 |
| *dirty | Yellow + Bold | 注意喚起 |
| エラー | Red | エラーメッセージ |
| 成功（✓） | Green | 完了メッセージ |
| パス | Cyan | ファイルパス |

#### 3.3.4 UX 原則

1. **メイン操作は最短で**: `git slot <slot> <branch>` の2引数だけで完結
2. **Fail Fast**: 不正な入力は即座にエラーを返し、修正方法を提示する
3. **Confirmation for Destructive Actions**: `--clear` の Dirty 状態は確認プロンプト。`--force` でスキップ可能
4. **Non-TTY Fallback**: パイプや CI 環境では色なしのプレーン出力
5. **cd 連携**: `git slot <slot>` のパス出力は stdout のみ（stderr にメッセージを出さない）

### 3.4 Error Handling

| エラー状況 | 出力例 |
|-----------|--------|
| 引数なし | Usage を表示 |
| 排他フラグ同時指定 | "Error: --list and --clear cannot be used together." |
| 不明なスロット名 | "Error: unknown slot 'foo'. Defined slots: main-work, hotfix, experiment" |
| 不明なフラグ | "Error: unknown flag '--xyz'. Run 'git slot --help'." |
| Empty スロットへのパス取得 | "Error: slot 'experiment' is empty. Load a branch first: git slot experiment <branch>" |

## 4. Phase / Priority

| 機能 | フェーズ | 優先度 |
|------|---------|--------|
| `git slot <slot> <branch>` (Load) | Phase 2 | P0 |
| `git slot <slot> -c <branch>` (Create + Load) | Phase 2 | P0 |
| `git slot <slot>` (パス出力) | Phase 2 | P0 |
| `git slot --list` | Phase 2 | P0 |
| `git slot --clear <slot>` | Phase 2 | P0 |
| `git slot --version` | Phase 2 | P0 |
| `git slot --swap <A> <B>` | Phase 2 | P1 |
| `git slot --status [slot]` | Phase 2 | P1 |
| `git slot --init` | Phase 2 | P1 |
| `--json` フラグ | Phase 2 | P2 |
| Non-TTY フォールバック | Phase 2 | P1 |
| `git slot completion` | Phase 3 | P2 |
