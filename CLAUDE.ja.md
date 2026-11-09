[繁體中文](CLAUDE.md) | [English](CLAUDE.en.md) | **日本語**

# CLAUDE.md — go-desktop-kit

Claude 向けのプロジェクト説明です。このリポジトリは初日から GitHub で公開しているため、公開しても問題のない内容だけを記載します。

このファイルは翻訳版です。主版は `CLAUDE.md`（繁体字中国語）で、Claude Code が自動的に読み込むのはそのファイルだけです。3 つの版は同じコミットで更新してください。

## 目的

Gio に依存しない再利用可能なパッケージ：`tray`（`fyne.io/systray` のラッパー）、`desktop/{single,autostart,paths,i18n,userauth}`、`secretstore`（keyring・agefile・memory の 3 つのバックエンド）、`sealbox`（パスワードによる暗号化、age 形式、パスワードは先に NFC 正規化）。GUI を使わないプログラムはこのリポジトリだけを import する。

最初の利用者は go-credential-porter。go-virtual-display も後でこれらのパッケージに切り替える。

## 設計方針

- パッケージは「トークン」や「仮想ディスプレイ」を知らない。業務上の文言はすべて呼び出し側から渡す。
- `internal/` は置かず、どのリポジトリからも import できるようにする。
- パッケージ名はすべて小文字の 1 単語。Gio 関連には `gio` の接頭辞を付け、Gio 自身のパッケージとの名前の衝突を避ける。
- このリポジトリ内のパッケージはまとめてリリースし、タグは通常の `v0.1.0` 形式。後でパッケージを独立した module に分けるときは、そのフォルダに go.mod を追加し、ambiguous import を避けるため直ちにルート module の新バージョンをリリースする。
- 機密内容を出力する API を追加しない。

## GitHub に関する決定

- **GitHub に関する決定**（2026-11-09）：公開の GitHub を主な開発場所とし（`github.com/The78estHerry/go-desktop-kit`）、Gitea はプルミラーにする。ミラー上ではコミットしない。
- 準備済み：MIT LICENSE（The78ester）、`.gitignore`。
- テストとサンプルに本物の資格情報、アカウント、内部 URL を入れない（`example.com`、`your-name`、`test-token-not-real` を使う）。

## プラットフォーム

- Windows ではすべて CGO 不要。`desktop/userauth` は macOS で CGO が必要（LocalAuthentication）。
- 現在は Windows でのみ実機テストを行う。macOS と Linux は「実装済み・未テスト」と表示し、プラットフォーム固有のファイルの先頭に `// Status: implemented, not yet tested on real hardware.` を書く。

## 言語の規則

- コードコメントと Godoc はすべて英語。
- PR の説明、issue の本文、コードレビューのコメントはすべて英語。
- コミットメッセージは 3 言語（英語が主、繁体字中国語が次、日本語が補足）。形式はグローバル CLAUDE.md の「Commit message language (all projects)」を参照。
- `README.md` が主版で繁体字中国語。`README.en.md`（英語）と `README.ja.md`（日本語）は同じコミットで同期する。`CLAUDE.md` の 3 つの版も同様。
- 開発者への返信は常に繁体字中国語。

## 進捗記録

- 2026-11-09：リポジトリの骨組みを作成（go.mod、LICENSE、.gitignore、3 言語版の CLAUDE.md）。
