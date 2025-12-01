# Interactive Deletion Mode

インタラクティブ削除モードは、重複ファイルを見つけた後に対話的に削除するファイルを選択できる機能です。

## 使い方

### 基本的な使用方法

```bash
# インタラクティブモードを有効にする
./dup-finder dir1 dir2 --interactive

# ハッシュを事前計算してから実行（推奨）
./dup-finder dir1 dir2 --compare-hash --interactive

# 複数ディレクトリで使用
./dup-finder dir1 dir2 dir3 --interactive
```

### オプション

- `--interactive` / `-i`: インタラクティブ削除モードを有効化
- `--compare-hash` / `-H`: ハッシュを事前計算（インタラクティブモードと併用推奨）

## 機能

### 1. 重複ファイルの表示

各重複セットについて以下の情報を表示：
- ファイルパス
- ファイルサイズ（人間が読みやすい形式）
- 更新日時
- xxHashハッシュ（最初の16文字、計算済みの場合のみ）

### 2. アクション選択

各重複セットに対して以下のアクションを選択できます：

- **[s] Skip**: 何もしない
- **[1] Delete file 1**: ファイル1を削除（ファイル2を残す）
- **[2] Delete file 2**: ファイル2を削除（ファイル1を残す）
- **[h] Compute hash**: ハッシュを計算してファイルが本当に同一かを確認（ハッシュ未計算時のみ）
- **[a] Keep all from dir1**: dir1の全てのファイルを残してdir2を削除（2ディレクトリ比較時のみ）
- **[b] Keep all from dir2**: dir2の全てのファイルを残してdir1を削除（2ディレクトリ比較時のみ）
- **[f] Finish**: 現在までの選択で確認画面に進む（残りの重複をスキップ）
- **[q] Quit**: インタラクティブモードを終了

### 3. バッチ削除モード

2つのディレクトリを比較している場合、`[a]`または`[b]`を選択することで、残りの全ての重複セットに同じルールを自動適用できます。

例：`[a]`を選択すると、以降の全ての重複でdir1のファイルが保持され、dir2のファイルが削除対象になります。

### 4. 最終確認

全ての選択が完了すると、削除予定のファイル一覧が表示されます：
- 削除されるファイルのリスト（サイズ付き）
- 合計で解放される容量
- 実行または中止の選択

### 5. 実行結果サマリー

削除が完了すると、以下の情報を含むサマリーが表示されます：
- 処理した重複セット数
- 削除されたファイル数
- 解放された容量
- 成功した削除の詳細
- 失敗した削除の詳細（エラー情報付き）

## 動作の流れ

```
1. 重複ファイルの検出
   ↓
2. ハッシュの計算（必要な場合）
   ↓
3. 各重複セットの表示と選択
   ↓
4. 最終確認画面
   ↓
5. ファイルの削除実行
   ↓
6. サマリーの表示
```

## ハッシュの計算

- `--compare-hash`を使用した場合：ハッシュは事前に計算済み
- `--compare-hash`を使用しない場合：
  - 各重複セットで `[h]` を選択してオンデマンドで計算可能
  - ファイルサイズが同じでも内容が異なる場合を検出できる

インタラクティブモードでは、ファイルの内容が本当に同一かを確認するため、ハッシュ計算を推奨します。

## 安全機能

### 削除前のチェック

- ファイルが存在することを確認
- 通常のファイルであることを確認（ディレクトリやシンボリックリンクではない）
- 親ディレクトリへの書き込み権限を確認

### エラーハンドリング

- 削除に失敗しても処理を継続
- 全てのエラーを収集してサマリーで表示
- 失敗したファイルは手動で対処可能

## 使用例

### 例1: 基本的な使い方

```bash
$ ./dup-finder ~/Downloads ~/Backup --compare-hash --interactive

=== /home/user/Downloads ↔ /home/user/Backup ===
photo.jpg:           ✓ [Hash: ✓ Identical]
document.pdf:        ✓ [Hash: ✓ Identical]

--- Entering Interactive Deletion Mode ---

=== Duplicate Set #1 ===
Found 2 files with same size
Hash: a1b2c3d4e5f6g7h8... (verified)

[1] /home/user/Downloads/photo.jpg
    Size: 2.3 MB
    Modified: 2024-11-20 14:30:00

[2] /home/user/Backup/photo.jpg
    Size: 2.3 MB
    Modified: 2024-11-15 10:15:00

Choose an action:
  [s] Skip (do nothing)
  [1] Delete: /home/user/Backup/photo.jpg
  [2] Delete: /home/user/Downloads/photo.jpg
  [a] Keep all from /home/user/Downloads, delete all from /home/user/Backup
  [b] Keep all from /home/user/Backup, delete all from /home/user/Downloads
  [q] Quit interactive mode
  [f] Finish selection and proceed to confirmation

Your choice: 1

...（他の重複ファイルの処理）

=== Final Confirmation ===
The following 2 file(s) will be deleted:

1. /home/user/Backup/photo.jpg (2.3 MB)
2. /home/user/Backup/document.pdf (1.5 MB)

Total space to be freed: 3.8 MB

Options:
  [y] Execute deletions (proceed)
  [n] Cancel all deletions (abort)

Your choice [y/N]: y

=== Interactive Session Summary ===
Duplicate Sets Found: 2
Files Deleted: 2
Space Freed: 3.8 MB

Successfully Deleted:
  ✓ /home/user/Backup/photo.jpg (2.3 MB freed)
  ✓ /home/user/Backup/document.pdf (1.5 MB freed)
```

### 例2: バッチ削除モード

```bash
$ ./dup-finder old-backup new-backup --compare-hash --interactive

（最初の重複セットで [a] を選択）

Batch mode enabled: All remaining duplicates from new-backup will be deleted.

（残りの重複セットは自動的に処理される）
```

## 注意事項

- 削除されたファイルは復元できません
- 重要なファイルを削除する前に必ずバックアップを取ってください
- 権限エラーが発生した場合は、適切な権限で実行してください
- 3つ以上のディレクトリを比較する場合、バッチ削除オプションは利用できません

## トラブルシューティング

### "No duplicate files found (based on content hash)"

→ 名前が同じでも内容が異なるファイルのみが見つかりました。これらは重複ファイルではありません。

### "permission denied"

→ ファイルまたは親ディレクトリへの書き込み権限がありません。適切な権限で実行してください。

### "cannot access file"

→ ファイルが削除または移動された可能性があります。再度スキャンを実行してください。

## テスト

```bash
# ユニットテストの実行
go test ./internal/interactive/... -v

# カバレッジの確認
go test ./internal/interactive/... -cover

# 全体のテスト
go test ./...
```
