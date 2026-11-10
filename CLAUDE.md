**繁體中文** | [English](CLAUDE.en.md) | [日本語](CLAUDE.ja.md)

# CLAUDE.md — go-desktop-kit

給 Claude 的專案說明。這個 repo 從第一天起就在 GitHub 公開，內容只寫公開也沒關係的事。

本檔是主要版本，Claude Code 只會自動讀取這個檔案；`CLAUDE.en.md`（英文）與 `CLAUDE.ja.md`（日文）是翻譯，修改本檔時在同一個 commit 裡一起更新。

## 目的

不依賴 Gio 的可重用套件：`tray`（包裝 `fyne.io/systray`）、`desktop/{single,autostart,paths,i18n,userauth,apppin}`、`secretstore`（keyring、agefile、memory 三個後端）、`sealbox`（以密碼加密，age 格式，密碼先做 NFC 正規化）。不用 GUI 的程式只 import 這個 repo。

第一個使用者是 go-credential-porter；go-virtual-display 之後也改用這裡的套件。

## 設計原則

- 套件不知道「Token」或「虛擬螢幕」是什麼，任何業務文字都由呼叫端傳入。
- 不放 `internal/`，讓其他 repo 都能 import。
- 套件名稱全小寫、單一單字；Gio 相關的加 `gio` 前綴，避免和 Gio 自己的套件撞名。
- 同一個 repo 內的套件一起發佈版本，標籤是普通的 `v0.1.0`。某個套件之後要拆成獨立 module 時，在該資料夾加 go.mod，並立刻從根 module 發新版，避免 ambiguous import。
- 不得新增會輸出機密內容的 API。

## GitHub 決定

- **GitHub 決定**（2026-11-09）：GitHub 公開為主要開發位置（`github.com/The78estHerry/go-desktop-kit`）；Gitea（`The78estWorld/go-desktop-kit`）也保留一份，兩邊都推：remote `origin` 是 Gitea、`github` 是 GitHub，分支與 tag 兩邊都推，Release 兩邊都建（步驟見 `RELEASING.md`）。
- 已準備：MIT LICENSE（The78ester）、`.gitignore`。
- 測試與範例不放真實憑證、帳號或內部網址（用 `example.com`、`your-name`、`test-token-not-real`）。

## 平台

- Windows 全程免 CGO；`desktop/userauth` 在 macOS 需要 CGO（LocalAuthentication）。
- 目前只在 Windows 實機測試；macOS、Linux 標示「已實作、未測試」：平台專屬檔案開頭加 `// Status: implemented, not yet tested on real hardware.`。

## 語言規則

- 程式碼註解與 Godoc 一律用英文。
- PR 說明、issue 內容、程式碼審查留言一律用英文。
- commit 訊息三語：英文為主、繁體中文次之、日文為輔，格式見全域 CLAUDE.md 的「Commit message language (all projects)」。
- `README.md` 是主要版本，用繁體中文；`README.en.md`（英文）與 `README.ja.md`（日文）在同一個 commit 裡同步更新。`CLAUDE.md` 三個版本也一樣。
- 回覆開發者一律用繁體中文。

## 進度紀錄

- 2026-11-09：建立 repo 骨架（go.mod、LICENSE、.gitignore、CLAUDE.md 三語版本）。
- 2026-11-09：M1 完成 `sealbox`、`secretstore`（keyring、agefile、memory）、`desktop/paths` 與單元測試（secretstore 實際寫入 Windows Credential Manager 測試，測完清除）；三平台 GitHub Actions；圖示定案（積木 B3：GO＋鎖＋資料夾＋系統列＋速度線，`assets/icon.svg`，`assets/make-icon.ps1` 轉 PNG）；三語 README。尚未公開、尚未打 v0.1.0 tag。
  - 已知：age 檔案後端每次寫入約 1 秒（整個檔案重新 scrypt 加密）；Windows 上 keyring 名稱是 `<appID>:<service>/<key>`（go-keyring 固定用 `:`）。
- 2026-11-09（暫停點）：公開 go-desktop-kit 進行到 gitea-to-github `checklist.md` 第 4 步。已完成：敏感資料檢查（全部乾淨）、使用者決定（歷史原封不動、CLAUDE.md 公開、新增 RELEASING.md、Gitea 兩邊都推）、Gitea repo 已建並推上 main。**下次從這裡接續**：請使用者在 GitHub 建空的 Public repo `The78estHerry/go-desktop-kit`（不勾任何初始檔），並把它加進 fine-grained token 的 Repository access（Contents: Read and write）；確認後 `git remote add github …`、推 main、打 `v0.1.0` tag 推兩邊、等 CI 通過、兩邊建 Release，最後在 gitea-to-github 寫 `cases/go-desktop-kit.md`。
- 2026-11-09：**v0.1.0 已發佈**。main 與 tag 推上 GitHub 與 Gitea，CI 三平台通過，兩邊都有 Release，module proxy 可取得。fine-grained token 需要 Contents 與 Workflows 兩個權限。
- 待辦：Gitea Actions 也會讀 `.github/workflows/ci.yml`，但 NAS 上只有 Linux runner，Windows、macOS 工作會一直排隊、不會執行（不會產生失敗通知）。之後考慮加 `.gitea/workflows/` 只跑 ubuntu，或接受現況。
- 2026-11-11：M4 第一步：新增 `desktop/single`（Windows 具名 mutex＋event；macOS／Linux／BSD 用 flock＋Unix socket，第二份啟動會請第一份顯示）與 `desktop/autostart`（Windows Run 機碼、macOS LaunchAgent、Linux XDG autostart；.desktop 與 plist 內容在核心產生，各平台只決定寫入位置）。Windows 實際測試（autostart 的測試寫入暫存機碼 `HKCU\Software\go-desktop-kit-test`，測完刪除，不碰真正的 Run 機碼）；Linux、macOS、FreeBSD 交叉 vet。尚未發版。下一步：`desktop/i18n`（訊息表，en／zh-TW／ja）。
- 2026-11-11：新增 `desktop/i18n`（重新設計，使用者決定）：`Catalog` 訊息表（缺字退回預設語言、再退回 key）、`Match` 依偏好語言挑選（完全相同 → 同語言同文字 → 同語言；只有繁中時所有中文使用者都用繁中）、`Missing` 供測試檢查翻譯是否齊全、`SystemLanguages`（Windows `GetUserPreferredUILanguages`、macOS `AppleLanguages`、其他看 gettext 環境變數）。舊版讀安裝設定的部分不搬（由 App 決定）。下一步：`tray`。
- 2026-11-11：新增 `tray`（包裝 `fyne.io/systray` v1.12.2）：`Start(ready, onExit)` 立即返回；選單用 callback（`AddItem`、`AddCheckbox`、子選單、`SetEnabled`／`SetVisible`／`SetChecked`／`SetIcon`／`Remove`），勾選項目不自動切換（動作成功後由呼叫端 `SetChecked`）；Windows 在 `LockOSThread` 的 goroutine 跑 `systray.Run`，PNG 自動包成 ICO；macOS 用 `RunWithExternalLoop` 掛在 GUI 的主迴圈（未測試）。`tray/example` 實機測試八項全部通過（使用者操作）。macOS 版需要 CGO，Windows 上不能交叉 vet，交給 CI。下一步：go-gio-kit 的 `giowidget`。
- 2026-11-11：M5：新增 `desktop/apppin`（argon2id、6–64 個英數字、連錯 5 次鎖 30 秒起加倍最多 1 小時，比對前先存次數）與 `desktop/userauth`（Windows Hello：搬 M0 spike 的 WinRT 程式，改在專用的鎖定執行緒執行、ctx 結束時取消；macOS LocalAuthentication 需要 CGO，未測試；Linux 回傳 Unsupported）。`Window` 是不透明型別，GUI 套件用 `NativeWindow` 建立，App 程式碼看不到 HWND。Hello 實機測試：可用性、跳出畫面、通過、取消都正確。giowin 產生 `userauth.Window` 的轉接留到 M6。
