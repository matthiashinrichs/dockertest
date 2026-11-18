Dockerbuild und CI/CD Test Projekt
==================================

## Container-Image

Der [GitHub-Actions-Workflow](.github/workflows/container.yml) baut das Image bei
Pushes auf `master` und kann auf diesem Branch auch manuell gestartet werden.
Er veröffentlicht es in der GitHub Container Registry (GHCR) unter
`ghcr.io/matthiashinrichs/dockertest`, mit der vollständigen Commit-SHA und
zusätzlich als `latest`. Der Workflow nutzt das automatisch bereitgestellte
`GITHUB_TOKEN`; ein separates Registry-Secret ist nicht nötig.

```sh
docker pull ghcr.io/matthiashinrichs/dockertest:latest
```

Bei privaten Paketen ist vorher ein `docker login ghcr.io` mit einem
Personal Access Token (classic) mit `read:packages` und Zugriff auf das Paket
erforderlich.

## Jenkins-Ablösung

Der neue Workflow ersetzt Build und Registry-Push. Kubernetes-Deployment und
Slack-Benachrichtigungen werden nicht mehr automatisch ausgeführt. Das
bisherige [Kubernetes-Manifest](myapp.yaml) verweist noch auf die alte Registry
und muss vor einem Deployment aus GHCR angepasst werden. Den bisherigen
Jenkins-Job nach erfolgreicher Umstellung separat deaktivieren.