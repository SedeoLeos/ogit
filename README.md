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

## Build Linux

### Archive zip directe

```powershell
pwsh ./scripts/build-zip.ps1
```

L'archive est generee dans `dist/ogit-linux-amd64.zip` et contient le binaire `ogit`.

### Binaire seul

```bash
./scripts/build-linux.sh
```

Le binaire est genere dans `dist/ogit-linux-amd64`.

### A la main

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/ogit-linux-amd64 .
```

## Installation sur le serveur

1. Copier `dist/ogit-linux-amd64.zip` sur le serveur.
2. Decompresser l'archive.
3. Installer `ogit` dans `/opt/ogit/ogit` ou dans un dossier de ton choix.
4. Placer la configuration dans `/opt/ogit/.env`.
5. Rendre le fichier executable.

Exemple:

```bash
sudo mkdir -p /opt/ogit
unzip ogit-linux-amd64.zip
sudo install -m 755 ogit /opt/ogit/ogit
sudo install -m 600 .env /opt/ogit/.env
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

## Arborescence ajoutee

- `scripts/build-linux.sh`
- `scripts/build-zip.ps1`
- `.github/workflows/ci.yml`
