# GitHub Release (Windows)

## What is implemented

- Push tag (`v*`) -> auto build + package + create GitHub Release.
- Manual trigger (`workflow_dispatch`) -> custom tag for v2 release.
- Each release publishes 2 packages:
  - `v2_full`: embedded frontend, double-click exe can open page.
  - `v2_backend`: backend-only executable.
- CI optimization enabled:
  - `concurrency` cancel-in-progress
  - pnpm cache via `actions/setup-node`
  - Go cache via `actions/setup-go`
  - Node `22`, pnpm `10`
  - Dependabot auto-merge for `semver-patch` / `semver-minor` (major excluded)

## Dist policy

- Frontend `dist` is not committed.
- Action builds frontend dist on each run and injects it into full package by build tag `with_frontend`.
- Backend package uses tag `no_frontend`.

Workflow file:
- `.github/workflows/release-windows.yml`

## Usage

### 1. Auto release by tag push

```bash
git tag v1.2.3
git push origin v1.2.3
```

After workflow success, GitHub Release will be created with zip package.

### 2. Manual release in GitHub Actions

1. Open `Actions` -> `Release Windows` -> `Run workflow`.
2. Fill:
   - `tag_name` (example `v1.2.3`)
   - `publish_release` (`true` to publish, `false` only artifact)

## Output

Release artifact naming:

`lol-shield_<tag>_windows_amd64_v2_full.zip`

`lol-shield_<tag>_windows_amd64_v2_backend.zip`

Example:

`lol-shield_v1.2.3_windows_amd64_v2_full.zip`

`lol-shield_v1.2.3_windows_amd64_v2_backend.zip`
