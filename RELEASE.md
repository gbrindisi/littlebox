# Releasing

Releases are automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. Pushing a semver tag triggers the pipeline.

## Cut a release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The `release.yml` workflow will build binaries for linux/darwin (amd64/arm64), generate checksums, and create a GitHub release with an auto-generated changelog.

## Dry run

Test the release locally without publishing:

```bash
goreleaser release --snapshot --clean
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/). The version string is injected at build time via `-ldflags` and is visible through:

```bash
littlebox --version
```

Local builds without a tag fall back to the short commit SHA (e.g. `47681b9-dirty`).
