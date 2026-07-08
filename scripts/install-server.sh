#!/usr/bin/env sh
set -eu

BIN_NAME="ogit"
INSTALL_DIR="/opt/ogit"
INSTALL_BIN="${INSTALL_DIR}/${BIN_NAME}"
SYSTEM_BIN="/usr/local/bin/${BIN_NAME}"
ENV_SAMPLE="./env.sample"

if [ ! -f "./${BIN_NAME}" ]; then
  if [ -f "./dist/${BIN_NAME}-linux-amd64" ]; then
    cp "./dist/${BIN_NAME}-linux-amd64" "./${BIN_NAME}"
  else
    echo "Missing binary. Build it first or place ./ogit next to this script." >&2
    exit 1
  fi
fi

sudo install -d "$INSTALL_DIR"
sudo install -m 755 "./${BIN_NAME}" "$INSTALL_BIN"
sudo ln -sf "$INSTALL_BIN" "$SYSTEM_BIN"

if [ -f "$ENV_SAMPLE" ] && [ ! -f "$INSTALL_DIR/.env" ]; then
  sudo install -m 600 "$ENV_SAMPLE" "$INSTALL_DIR/.env"
  echo "Created $INSTALL_DIR/.env from env.sample"
else
  echo "$INSTALL_DIR/.env already exists or env.sample missing"
fi

echo "Installed $BIN_NAME to $INSTALL_BIN and linked $SYSTEM_BIN"
echo "Edit $INSTALL_DIR/.env with your GitHub App credentials"
