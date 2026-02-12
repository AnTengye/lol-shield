# GitHub Release (Windows)

## What is implemented

- Push tag (`v*`) -> auto build + package + create GitHub Release.
- Manual trigger (`workflow_dispatch`) -> custom tag for v2 release.

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

`lol-shield_<tag>_windows_amd64_v2.zip`

Example:

`lol-shield_v1.2.3_windows_amd64_v2.zip`
