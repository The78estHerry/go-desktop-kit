[繁體中文](CLAUDE.md) | [English](CLAUDE.en.md) | **日本語**

# CLAUDE.md — go-desktop-kit

Claude 向けのプロジェクト説明です。このリポジトリは初日から GitHub で公開しているため、公開しても問題のない内容だけを記載します。

このファイルは翻訳版です。主版は `CLAUDE.md`（繁体字中国語）で、Claude Code が自動的に読み込むのはそのファイルだけです。3 つの版は同じコミットで更新してください。

## 目的

Gio に依存しない再利用可能なパッケージ：`tray`（`fyne.io/systray` のラッパー）、`desktop/{single,autostart,paths,i18n,userauth,apppin}`、`secretstore`（keyring・agefile・memory の 3 つのバックエンド）、`sealbox`（パスワードによる暗号化、age 形式、パスワードは先に NFC 正規化）。GUI を使わないプログラムはこのリポジトリだけを import する。

最初の利用者は go-credential-porter。go-virtual-display も後でこれらのパッケージに切り替える。

## 設計方針

- パッケージは「トークン」や「仮想ディスプレイ」を知らない。業務上の文言はすべて呼び出し側から渡す。
- `internal/` は置かず、どのリポジトリからも import できるようにする。
- パッケージ名はすべて小文字の 1 単語。Gio 関連には `gio` の接頭辞を付け、Gio 自身のパッケージとの名前の衝突を避ける。
- このリポジトリ内のパッケージはまとめてリリースし、タグは通常の `v0.1.0` 形式。後でパッケージを独立した module に分けるときは、そのフォルダに go.mod を追加し、ambiguous import を避けるため直ちにルート module の新バージョンをリリースする。
- 機密内容を出力する API を追加しない。

## GitHub に関する決定

- **GitHub に関する決定**（2026-11-09）：公開の GitHub を主な開発場所とし（`github.com/The78estHerry/go-desktop-kit`）、Gitea（`The78estWorld/go-desktop-kit`）にも置いて両方にプッシュする。remote は `origin` が Gitea、`github` が GitHub。ブランチとタグは両方にプッシュし、リリースも両方で作成する（手順は `RELEASING.md`）。
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
- 2026-11-09：M1 完了。`sealbox`、`secretstore`（keyring・agefile・memory）、`desktop/paths` と単体テスト（secretstore は実際の Windows 資格情報マネージャーでテストし、終了後に削除）、3 プラットフォームの GitHub Actions、アイコン決定（積み木 B3：GO・錠前・フォルダー・トレイ・スピード線、`assets/icon.svg`、`assets/make-icon.ps1` で PNG に変換）、3 言語の README。未公開、v0.1.0 タグは未作成。
  - 既知の点：age ファイルのバックエンドは書き込みごとに約 1 秒かかる（ファイル全体を scrypt で再暗号化）。Windows での keyring 名は `<appID>:<service>/<key>`（go-keyring は常に `:` を使う）。
- 2026-11-09（一時停止）：go-desktop-kit の公開は gitea-to-github の `checklist.md` 第 4 段階まで進んだ。完了：機密データの確認（すべて問題なし）、ユーザーの決定（履歴はそのまま、CLAUDE.md は公開、RELEASING.md を追加、GitHub と Gitea の両方にプッシュ）、Gitea リポジトリの作成と main のプッシュ。**ここから再開**：ユーザーに GitHub で空の公開リポジトリ `The78estHerry/go-desktop-kit`（初期ファイルなし）を作成し、fine-grained token の Repository access（Contents: Read and write）に追加してもらう。その後 `git remote add github …`、main をプッシュ、`v0.1.0` タグを両方にプッシュ、CI の通過を待ち、両方でリリースを作成し、最後に gitea-to-github に `cases/go-desktop-kit.md` を書く。
- 2026-11-09：**v0.1.0 をリリース**。main とタグを GitHub と Gitea にプッシュし、CI は 3 プラットフォームで合格、両方にリリースを作成、module proxy から取得できる。fine-grained token には Contents と Workflows の両方の権限が必要。
- TODO：Gitea Actions も `.github/workflows/ci.yml` を読むが、NAS には Linux の runner しかないため、Windows と macOS のジョブはキューに残ったまま実行されない（失敗通知は出ない）。ubuntu だけを実行する `.gitea/workflows/` を追加するか、現状のままにするか検討する。
- 2026-11-11：M4 の第 1 段階：`desktop/single`（Windows は名前付き mutex＋event、macOS／Linux／BSD は flock＋Unix ソケット、2 つ目の起動で 1 つ目に表示を依頼）と `desktop/autostart`（Windows の Run キー、macOS の LaunchAgent、Linux の XDG autostart。.desktop と plist の内容はコアで生成し、各プラットフォームは書き込み先だけを決める）を追加。Windows で実テスト（autostart のテストは一時キー `HKCU\Software\go-desktop-kit-test` に書き込み、終了後に削除。本物の Run キーには触れない）、Linux・macOS・FreeBSD はクロス vet。未リリース。次：`desktop/i18n`（メッセージ表、en／zh-TW／ja）。
- 2026-11-11：`desktop/i18n` を追加（ユーザーの決定で作り直し）：`Catalog` メッセージ表（既定の言語、次に key にフォールバック）、`Match` は優先言語から選択（完全一致 → 同じ言語と文字 → 同じ言語。繁体字しかなければ中国語の読者は全員繁体字）、`Missing` でテストから訳の抜けを確認、`SystemLanguages`（Windows は `GetUserPreferredUILanguages`、macOS は `AppleLanguages`、その他は gettext の環境変数）。旧版のインストール設定の読み取りは移さない（アプリが決める）。次：`tray`。
- 2026-11-11：`tray` を追加（`fyne.io/systray` v1.12.2 を包む）：`Start(ready, onExit)` はすぐ戻る。メニューはコールバック（`AddItem`、`AddCheckbox`、サブメニュー、`SetEnabled`／`SetVisible`／`SetChecked`／`SetIcon`／`Remove`）。チェック項目は自動で切り替わらない（操作が成功したら呼び出し側が `SetChecked`）。Windows は `LockOSThread` した goroutine で `systray.Run`、PNG は ICO に包む。macOS は `RunWithExternalLoop` で GUI のメインループに載せる（未テスト）。`tray/example` の実機テスト 8 項目すべて合格（ユーザー操作）。macOS 版は CGO が必要で Windows からクロス vet できないため CI に任せる。次：go-gio-kit の `giowidget`。
