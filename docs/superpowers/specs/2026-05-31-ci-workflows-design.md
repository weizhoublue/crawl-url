# CI Workflows Design

This specification defines the GitHub Actions CI/CD workflows for the `crawl-url` project.

## 1. Workflows Overview

We implement three separate workflow files under `.github/workflows/` to support release automation, Pull Request verification, and manual artifact builds.

---

## 2. Workflows Specification

### 2.1 Release CI (`release.yml`)
* **Trigger**: Pushing a tag matching pattern `v*` (e.g., `v1.0.0`).
* **Permissions**:
  ```yaml
  permissions:
    contents: write
  ```
* **Steps**:
  1. **Checkout**: Check out code for the triggered tag.
  2. **Go Setup**: Setup Go environment using version from `go.mod`.
  3. **Test**: Run `go test ./...`.
  4. **Build**: Build binaries for three platforms:
     - Linux AMD64 (`GOOS=linux GOARCH=amd64`)
     - Linux ARM64 (`GOOS=linux GOARCH=arm64`)
     - macOS ARM64 (`GOOS=darwin GOARCH=arm64`)
  5. **Create Release**: Draft and publish a GitHub Release with the three compiled binaries attached.
  6. **Create Release Branch**: Create and push a branch named `release/vx.y.z` matching the tag name.

### 2.2 PR CI (`pr.yml`)
* **Trigger**: Pull Request targeting the `main` branch.
* **Steps**:
  1. **Checkout**: Check out the PR branch code.
  2. **Go Setup**: Setup Go environment.
  3. **Test**: Run `go test ./...` (runs all unit and E2E tests).
  4. **Build Check**: Test compile using `go build ./cmd/crawl-url`.

### 2.3 Build CI (`build.yml`)
* **Trigger**: Manual invocation (`workflow_dispatch`).
* **Inputs**:
  * `ref`: The Git branch name, commit hash, or tag to build from (required).
* **Steps**:
  1. **Checkout**: Check out the code corresponding to the specified `ref`.
  2. **Go Setup**: Setup Go environment.
  3. **Test**: Run `go test ./...`.
  4. **Build**: Build binaries for the three platforms (Linux AMD64, Linux ARM64, macOS ARM64).
  5. **Upload Artifacts**: Upload build artifacts to GitHub Action run history.

---

## 3. Success Criteria
* Tag pushes successfully trigger `release.yml`, compiling binaries, creating release, and pushing branch `release/vx.y.z`.
* Pull Requests to `main` trigger `pr.yml` and successfully run all tests and compile checks.
* Manual triggers with valid refs (branches, tags, commits) trigger `build.yml` and upload compiled binaries.
