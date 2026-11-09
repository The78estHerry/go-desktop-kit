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
