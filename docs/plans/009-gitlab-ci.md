# Plan 009: GitLab CI/CD Pipeline

## Context
The project had no CI/CD configuration. Goal: Automated tests with coverage, Goreleaser builds, and Docker images in GitLab CI.

## Implemented Changes

### `.gitlab-ci.yml`
- **Stages:** `test` → `build` → `docker`
- **test** (all branches): Unit + integration tests with coverage, HTML report as artifact
- **build:goreleaser** (master only): Tag → release, no tag → snapshot
- **build:manual** (non-master, manual): `make build` + `make build-dev`
- **docker** (master only): Kaniko-based image build, push to `$CI_REGISTRY_IMAGE`

### `.goreleaser.yml`
- Build for linux/amd64, linux/arm64, darwin/arm64, windows/amd64
- Ldflags with buildinfo variables (Version, Commit, BuildDate, GoVersion)
- tar.gz for Linux/macOS, zip for Windows
- GitLab release integration on tags

### `.dockerignore`
- Excludes `.git`, `bin/`, `dist/`, DB files, `.claude/`

### `Dockerfile`
- ARGs for VERSION, COMMIT, BUILD_DATE with fallback values
- Ldflags inject buildinfo variables in Docker build

## Verification
```bash
goreleaser check          # Validate Goreleaser config
make test                 # Tests green locally
git push origin master    # Check pipeline in GitLab
```
