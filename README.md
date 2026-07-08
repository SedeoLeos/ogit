# ogit

Binaire Go à compiler puis installer sur un serveur Linux.

## Prerequis

- Go 1.22 ou plus recent
- Un serveur Linux pour executer le binaire

## Developpement local

```bash
go mod tidy
go run .
```

Pour un env local, tu peux definir:

```bash
set OGIT_ENV_FILE=.env
```

## Lint

```bash
make lint
```

Le lint exécute:

- `gofmt -l .`
- `go vet ./...`

## Build Linux

### Archive zip directe

```powershell
pwsh ./scripts/build-zip.ps1
```

L'archive est generee dans `dist/ogit-linux-amd64.zip` et contient le binaire `ogit`.

### Binaire seul

```bash
make build
```

Le binaire est genere dans `dist/ogit-linux-amd64`.

### A la main

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ogit-linux-amd64 .
```

## Installation sur le serveur

Le plus simple est:

```bash
make install-server
```

Ce script:

- crée `/opt/ogit`
- installe le binaire dans `/opt/ogit/ogit`
- crée le lien `/usr/local/bin/ogit`
- crée `/opt/ogit/.env` depuis `env.sample` si le fichier n'existe pas encore

Ensuite tu peux éditer le fichier:

```bash
sudo nano /opt/ogit/.env
```

Exemple de contenu:

```env
GITHUB_INSTALLATION_ID=123456
GITHUB_APP_ID=123456
GITHUB_PRIVATE_KEY_PATH=/opt/ogit/private-key.pem
```

Puis copie la clé privée:

```bash
sudo install -m 600 private-key.pem /opt/ogit/private-key.pem
```

## Fichier .env

Le binaire charge par defaut `/opt/ogit/.env`.

Tu peux aussi surcharger le chemin avec:

```bash
export OGIT_ENV_FILE=/chemin/vers/.env
```

## Workflow CI

Le fichier `.github/workflows/ci.yml` lance:

- `go test ./...`
- un build Linux amd64 au format zip

## Commandes Git

- `ogit clone <url> [dossier]` transmet le dossier cible à `git clone`.
- `ogit pull [options]` transmet les options à `git pull`.
- `ogit fetch [options]` transmet les options à `git fetch`.
