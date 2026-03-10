# Releasing Specwatch

Specwatch releases are built with GoReleaser and published from GitHub Actions.

## What gets published

- GitHub release archives for Linux, macOS, and Windows
- `checksums.txt`
- Homebrew formula updates in `RajeshShrirao/homebrew-tap`
- Scoop manifest updates in `RajeshShrirao/scoop-bucket`

## Required repositories

Create these repositories under the same GitHub account before the first tagged release:

- `RajeshShrirao/homebrew-tap`
- `RajeshShrirao/scoop-bucket`

Repository layout:

- `homebrew-tap`: formula goes in `Formula/specwatch.rb`
- `scoop-bucket`: manifest goes in `bucket/specwatch.json`

The bootstrap files in `packaging/homebrew/Formula/specwatch.rb` and
`packaging/scoop/specwatch.json` show the expected structure for the first commit.

## Required secrets

Add this repository secret in GitHub Actions:

- `RELEASE_GITHUB_TOKEN`: GitHub personal access token with `repo` scope for pushing to:
  - `RajeshShrirao/Specwatch`
  - `RajeshShrirao/homebrew-tap`
  - `RajeshShrirao/scoop-bucket`

`GITHUB_TOKEN` handles the release in the current repository. `RELEASE_GITHUB_TOKEN` is used by
GoReleaser for the cross-repository Homebrew and Scoop updates.

## Release flow

1. Make sure CI is green on `main`.
2. Create and push a semver tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

3. GitHub Actions runs `.github/workflows/release.yml`.
4. GoReleaser builds archives, writes `checksums.txt`, creates the GitHub release, and updates:
   - `Formula/specwatch.rb` in `homebrew-tap`
   - `bucket/specwatch.json` in `scoop-bucket`

## Local dry run

Install GoReleaser, then run:

```bash
goreleaser release --snapshot --clean
```

This validates the config and produces local artifacts without publishing a release.
