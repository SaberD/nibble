# go-cli-release-template

A template for Go CLI tools with multi-channel distribution out of the box.

## What's included

- **Dev container**, ready to use Go development environment with [VHS](https://github.com/charmbracelet/vhs) for recording terminal demos. Edit `demo.tape` and run `make demo` to generate `demo.gif`
- **GoReleaser**, cross-platform builds (Linux, macOS, Windows x amd64/arm64)
- **releases**
   - **go install**, works automatically from GitHub tags
   - **Homebrew**, tap formula auto-updated on release
   - **npm**, global install via `npm install -g`
   - **pipx**, install via `pipx install`
- **Checksum verification**, all package installers are setup verify SHA256 against GoReleaser's checksums file
- **Smoke tests**, CI verifies the binary works for every install method
- **`make fix`**, runs `go fmt`, `go vet`, and the new [`go fix`](https://go.dev/blog/gofix) to automatically modernize your Go code (Go 1.26+)

## Setup

1. Click **Use this template** on GitHub to create your repo
2. Clone it and run the setup script:

```bash
go run setup.go
```

> all instances of `YOUR_OWNER` and `YOUR_PROJECT` in the repo are replaced by setup.go

3. Set up required [GitHub secrets and environments](https://github.com/YOUR_OWNER/YOUR_PROJECT/settings/secrets/actions):
   - `HOMEBREW_TAP_GITHUB_TOKEN`, token with write access to your tap repo
   - **npm**: create an npm org, then do a first manual publish to register the package name:
     ```bash
     cd npm-package && npm publish --access public
     ```
     After that, add your npm automation token as `NODE_AUTH_TOKEN` in the repo's `npm` GitHub environment. Run `make npm` locally to verify the package builds.
   - **PyPI**: use PyPI's [Pending Publisher](https://docs.pypi.org/trusted-publishers/creating-a-project-through-oidc/) to set up trusted publishing before the package exists, no manual first publish needed. No token required, PyPI authenticates via OIDC automatically. Run `make pip` locally to verify the package builds.

4. Create your Homebrew tap repo (`github.com/YOUR_OWNER/homebrew-tap`)

5. Build your CLI in `main.go`

## Releasing

Make a release on GitHub with a tag starting with `v` (e.g. `v0.1.0`). GoReleaser builds all platform binaries, creates the GitHub release, and updates the Homebrew tap automatically.

To publish to npm or PyPI, manually trigger those workflows from the [Actions tab](https://github.com/YOUR_OWNER/YOUR_PROJECT/actions) and enter the release tag (e.g. `v0.1.0`). This cannot be automated without additional setup, npm and PyPI each require their own credentials and release environments. Once a publish workflow succeeds, its install smoke test triggers automatically. The `go install` test is also manual, run it after the Go module proxy has indexed the new tag (can take a few minutes). Check with:

```bash
go list -m github.com/YOUR_OWNER/YOUR_PROJECT@latest
```

## Distribution

After setup, users can install via:

```bash
brew install YOUR_OWNER/tap/YOUR_PROJECT
npm install -g @YOUR_OWNER/YOUR_PROJECT
pipx install YOUR_PROJECT-cli
go install github.com/YOUR_OWNER/YOUR_PROJECT@latest
```

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/YOUR_OWNER/YOUR_PROJECT)