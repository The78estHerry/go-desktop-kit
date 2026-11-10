[繁體中文](CLAUDE.md) | **English** | [日本語](CLAUDE.ja.md)

# CLAUDE.md — go-desktop-kit

Project notes for Claude. This repository is public on GitHub from day one, so it only contains information that is safe to publish.

This file is a translation. `CLAUDE.md` (Traditional Chinese) is the primary version and the only one Claude Code loads automatically; update all three versions in the same commit.

## Purpose

Reusable packages without Gio: `tray` (wraps `fyne.io/systray`), `desktop/{single,autostart,paths,i18n,userauth,apppin}`, `secretstore` (keyring, agefile and memory backends) and `sealbox` (passphrase encryption in age format; passphrases are NFC-normalized first). Programs without a GUI import only this repository.

The first user is go-credential-porter; go-virtual-display will switch to these packages later.

## Design principles

- Packages know nothing about "tokens" or "virtual displays"; all domain text is passed in by the caller.
- No `internal/` directory, so any repository can import the packages.
- Package names are lowercase single words; Gio-related ones get a `gio` prefix to avoid clashing with Gio's own packages.
- All packages in this repository are released together with plain tags such as `v0.1.0`. To split a package into its own module later, add a go.mod in its directory and immediately release a new version of the root module to avoid ambiguous imports.
- Never add an API that outputs secret content.

## GitHub decision

- **GitHub decision** (2026-11-09): GitHub (public) is the primary development location (`github.com/The78estHerry/go-desktop-kit`); a copy is also kept on Gitea (`The78estWorld/go-desktop-kit`) and every push goes to both: remote `origin` is Gitea and `github` is GitHub; push branches and tags to both and create releases on both (steps in `RELEASING.md`).
- Prepared: MIT LICENSE (The78ester), `.gitignore`.
- No real credentials, accounts or internal URLs in tests and examples (use `example.com`, `your-name`, `test-token-not-real`).

## Platforms

- Windows builds need no CGO; `desktop/userauth` needs CGO on macOS (LocalAuthentication).
- Only Windows is tested on real hardware for now; macOS and Linux are marked "implemented, not yet tested": platform-specific files start with `// Status: implemented, not yet tested on real hardware.`

## Language rules

- Write all code comments and Godoc in English.
- Write all PR descriptions, issue text and review comments in English.
- Commit messages are trilingual (English primary, Traditional Chinese secondary, Japanese supplementary); see "Commit message language (all projects)" in the global CLAUDE.md.
- `README.md` is the primary README, written in Traditional Chinese. Keep `README.en.md` (English) and `README.ja.md` (Japanese) in sync in the same commit. The same applies to the three versions of `CLAUDE.md`.
- Always reply to the developer in Traditional Chinese.

## Progress log

- 2026-11-09: Created the repository skeleton (go.mod, LICENSE, .gitignore, CLAUDE.md in three languages).
- 2026-11-09: M1 done: `sealbox`, `secretstore` (keyring, agefile, memory) and `desktop/paths` with unit tests (secretstore is tested against the real Windows Credential Manager and cleaned up afterwards); GitHub Actions on three platforms; icon finalized (blocks B3: GO + padlock + folder + tray + speed lines, `assets/icon.svg`, rendered by `assets/make-icon.ps1`); READMEs in three languages. Not public yet, no v0.1.0 tag yet.
  - Known: each write of the age file backend takes about 1 second (the whole file is re-encrypted with scrypt); on Windows the keyring name is `<appID>:<service>/<key>` (go-keyring always uses `:`).
- 2026-11-09 (paused): publishing go-desktop-kit reached step 4 of gitea-to-github `checklist.md`. Done: sensitive data check (all clean), user decisions (keep history, publish CLAUDE.md, add RELEASING.md, push to both GitHub and Gitea), Gitea repo created and main pushed. **Resume here**: ask the user to create an empty public GitHub repo `The78estHerry/go-desktop-kit` (no initial files) and add it to the fine-grained token's repository access (Contents: Read and write); then `git remote add github …`, push main, tag `v0.1.0` and push to both, wait for CI, create releases on both sides, and finally write `cases/go-desktop-kit.md` in gitea-to-github.
- 2026-11-09: **v0.1.0 released**. main and the tag are on GitHub and Gitea, CI passes on all three platforms, both sides have a release, and the module proxy serves it. The fine-grained token needs both the Contents and Workflows permissions.
- To do: Gitea Actions also reads `.github/workflows/ci.yml`, but the NAS only has a Linux runner, so the Windows and macOS jobs stay queued and never run (no failure notifications). Consider a `.gitea/workflows/` file that runs ubuntu only, or accept it.
- 2026-11-11: M4 step 1: added `desktop/single` (Windows named mutex + event; macOS/Linux/BSD flock + Unix socket, a second launch asks the first to show itself) and `desktop/autostart` (Windows Run key, macOS LaunchAgent, Linux XDG autostart; the .desktop and plist content is generated in the core, platforms only choose where to write). Tested on Windows (the autostart test writes a scratch key `HKCU\Software\go-desktop-kit-test`, deleted afterwards, never the real Run key); cross-vet for Linux, macOS and FreeBSD. Not released yet. Next: `desktop/i18n` (message table, en/zh-TW/ja).
- 2026-11-11: added `desktop/i18n` (redesigned, per the user): `Catalog` message table (falls back to the default language, then the key), `Match` picks from the preferred languages (exact → same language and script → same language; with only Traditional Chinese, every Chinese reader gets it), `Missing` lets tests check translations are complete, `SystemLanguages` (Windows `GetUserPreferredUILanguages`, macOS `AppleLanguages`, gettext environment variables elsewhere). The old installer-setting lookup was not moved (that is up to each app). Next: `tray`.
- 2026-11-11: added `tray` (wraps `fyne.io/systray` v1.12.2): `Start(ready, onExit)` returns at once; the menu uses callbacks (`AddItem`, `AddCheckbox`, submenus, `SetEnabled`/`SetVisible`/`SetChecked`/`SetIcon`/`Remove`); checkboxes do not toggle themselves (the caller calls `SetChecked` once the action succeeds); on Windows `systray.Run` runs in a `LockOSThread` goroutine and PNG is wrapped as ICO; macOS uses `RunWithExternalLoop` on the GUI main loop (untested). `tray/example` passed all eight real-hardware checks (driven by the user). The macOS build needs CGO and cannot be cross-vetted on Windows; CI covers it. Next: `giowidget` in go-gio-kit.
