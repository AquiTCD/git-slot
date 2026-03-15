---
description: Bump version, update CHANGELOG, tag, and push to trigger GoReleaser
---
# /bump-version Workflow

This workflow automates the release process for git-slot. It determines the appropriate version bump, updates the CHANGELOG, commits, tags, and pushes — triggering the GoReleaser CI pipeline.

## Execution Steps

1. **Restore Context**:
   - Run `git describe --tags --abbrev=0` to get the last tag (`LAST_TAG`).
   - Run `git log ${LAST_TAG}..HEAD --oneline` to review **all** unreleased commits since the last tag.
   - Read `CHANGELOG.md` to identify the current latest version and formatting conventions.

2. **Determine Semantic Version Bump**:
   - Analyze `git log <last_tag>..HEAD --oneline` to list unreleased commits.
   - Based on Conventional Commits prefixes and SemVer rules, autonomously decide:
     - `major`: breaking changes (`feat!:`, `BREAKING CHANGE`)
     - `minor`: new features (`feat:`)
     - `patch`: fixes and improvements (`fix:`, `build:`, `refactor:`)
   - Declare the `NEW_VERSION` (e.g., `0.2.0`).

3. **Update CHANGELOG.md**:
   - Read full `CHANGELOG.md`.
   - Insert a new section **below the header** (after line 6, before the first `## [x.y.z]` entry) in the format:
     ```
     ## [NEW_VERSION] - YYYY-MM-DD
     ```
   - Group changes into `### Added`, `### Changed`, and `### Fixed` as appropriate.
   - Add a link reference at the bottom: `[NEW_VERSION]: https://github.com/AquiTCD/git-slot/releases/tag/vNEW_VERSION`

// turbo
4. **Run Tests**:
   - Run `go test -race ./...` to confirm nothing is broken.
   - Run `go vet ./...` for static analysis.

5. **Stage and Commit**:
   - Verify that git status only shows `CHANGELOG.md` as modified.
   - Run: `git add CHANGELOG.md && git commit -m "chore(release): bump version to vNEW_VERSION"`
   - Adapt the commit body with a brief summary of the changelog.

6. **Tag and Push**:
   - Run: `git tag vNEW_VERSION`
   - Run: `git push origin main --tags`
   - This automatically triggers the `.github/workflows/release.yml` (GoReleaser).

7. **Confirmation**:
   - Notify the user that:
     - Version has been bumped to `vNEW_VERSION`
     - Tag has been pushed
     - GoReleaser CI is now running (link to Actions tab)
     - Users will be able to install via `brew tap AquiTCD/tap && brew install git-slot`
