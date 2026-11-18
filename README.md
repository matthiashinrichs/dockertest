# Dockertest

## Container-Image

Der [GitHub-Actions-Workflow](.github/workflows/container.yml) baut das Image
für `linux/amd64` und `linux/arm64` bei Pushes auf `master` und kann auf diesem
Branch auch manuell gestartet werden. Er veröffentlicht es in der GitHub
Container Registry (GHCR) unter
`ghcr.io/matthiashinrichs/dockertest`, mit der vollständigen Commit-SHA und
zusätzlich als `latest`. Beide Tags enthalten beide Architekturen; Docker
wählt beim Pull automatisch die passende Variante. Der Workflow nutzt das
automatisch bereitgestellte `GITHUB_TOKEN`; ein separates Registry-Secret ist
nicht nötig.

```sh
docker pull ghcr.io/matthiashinrichs/dockertest:latest
```

Bei privaten Paketen ist vorher ein `docker login ghcr.io` mit einem
Personal Access Token (classic) mit `read:packages` und Zugriff auf das Paket
erforderlich.
