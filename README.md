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
3. Rendre `ogit` executable si besoin.
4. Le placer dans un dossier de ton choix, par exemple `/opt/ogit/ogit`.

Exemple:

```bash
sudo install -d /opt/ogit
unzip ogit-linux-amd64.zip
sudo install -m 755 ogit /opt/ogit/ogit
```

## Workflow CI

Le fichier `.github/workflows/ci.yml` lance:

- `go test ./...`
- un build Linux amd64 au format zip

## Arborescence ajoutee

- `scripts/build-linux.sh`
- `scripts/build-zip.ps1`
- `.github/workflows/ci.yml`
