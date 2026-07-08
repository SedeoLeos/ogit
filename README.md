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

## Installation sur le serveur

1. Decompresser le zip.
2. Lancer:

```bash
sudo ./ogit config init
```

3. Renseigner `GITHUB_APP_ID` et `GITHUB_INSTALLATION_ID`.
4. Coller la clé privée PEM, puis taper `END` sur une ligne seule.
5. Le binaire est copie dans `/opt/ogit/ogit`.
6. Le lien `/usr/local/bin/ogit` est créé automatiquement.
7. Le fichier `/opt/ogit/.env` est créé automatiquement.
8. La clé privée est stockée dans `/opt/ogit/private-key.pem`.

Tu peux ensuite executer:

```bash
ogit clone <url> [dossier]
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
- `ogit config init` crée la configuration serveur et installe le binaire.
- `ogit config show` affiche le chemin du fichier env utilisé.
