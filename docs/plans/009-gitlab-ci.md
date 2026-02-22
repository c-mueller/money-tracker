# Plan 009: GitLab CI/CD Pipeline

## Kontext
Das Projekt hatte keine CI/CD-Konfiguration. Ziel: Tests mit Coverage, Goreleaser-Builds und Docker-Images automatisiert in GitLab CI.

## Umgesetzte Änderungen

### `.gitlab-ci.yml`
- **Stages:** `test` → `build` → `docker`
- **test** (alle Branches): Unit- + Integration-Tests mit Coverage, HTML-Report als Artifact
- **build:goreleaser** (nur master): Tag → Release, kein Tag → Snapshot
- **build:manual** (nicht-master, manuell): `make build` + `make build-dev`
- **docker** (nur master): Kaniko-basierter Image-Build, Push nach `$CI_REGISTRY_IMAGE`

### `.goreleaser.yml`
- Build für linux/amd64, linux/arm64, darwin/arm64, windows/amd64
- Ldflags mit buildinfo-Variablen (Version, Commit, BuildDate, GoVersion)
- tar.gz für Linux/macOS, zip für Windows
- GitLab Release-Integration bei Tags

### `.dockerignore`
- Excludiert `.git`, `bin/`, `dist/`, DB-Dateien, `.claude/`

### `Dockerfile`
- ARGs für VERSION, COMMIT, BUILD_DATE mit Fallback-Werten
- Ldflags injizieren buildinfo-Variablen im Docker-Build

## Verifikation
```bash
goreleaser check          # Goreleaser-Config validieren
make test                 # Tests lokal grün
git push origin master    # Pipeline in GitLab prüfen
```
