#!/bin/sh
set -e

# Usage:
#   curl -sL https://raw.githubusercontent.com/hmontazeri/mushak/main/install.sh | sh

OWNER="hmontazeri"
REPO="mushak"
BINARY="mushak"
INSTALL_DIR="/usr/local/bin"

get_latest_release() {
  curl --silent "https://api.github.com/repos/$OWNER/$REPO/releases/latest" | 
    grep '"tag_name":' | 
    sed -E 's/.*"([^"]+)".*/\1/'
}

download_release() {
  VERSION=$1
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  if [ "$ARCH" = "x86_64" ]; then
    ARCH="x86_64"
  elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
  fi

  # Match the naming convention in .goreleaser.yaml
  # Name template: {{ .ProjectName }}_{{ .Os }}_{{ .Arch }}
  # Example: mushak_Darwin_x86_64.tar.gz -> but wait, I used "title .Os" in goreleaser.yaml which makes it "Darwin" key, but often goreleaser defaults or my config produces lowercase usually or titled?
  # Let me double check standard goreleaser output or what I defined.
  # I defined: {{- title .Os }} -> so "Linux", "Darwin", "Windows"
  
  # Wait, standard uname output:
  # Linux -> Linux
  # Darwin -> Darwin
  
  # Let's adjust the script to match "Darwin", "Linux" to be safe.
  OS_TITLE=$(uname -s) # "Darwin" or "Linux"
  # But "Darwin" is usually what uname -s gives directly.
  
  # My .goreleaser.yaml:
  # name_template: >-
  #   {{ .ProjectName }}_
  #   {{- title .Os }}_
  
  FULL_ASSET_NAME="${BINARY}_${OS_TITLE}_${ARCH}.tar.gz"
  
  DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$VERSION/$FULL_ASSET_NAME"
  
  echo "Downloading $DOWNLOAD_URL..."
  curl -sL "$DOWNLOAD_URL" -o "/tmp/$FULL_ASSET_NAME"
  
  echo "Extracting..."
  tar -xzf "/tmp/$FULL_ASSET_NAME" -C /tmp
  
  echo "Installing to $INSTALL_DIR..."
  sudo mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"
  chmod +x "$INSTALL_DIR/$BINARY"
  
  echo "Clean up..."
  rm "/tmp/$FULL_ASSET_NAME"
}

main() {
  echo "Fetching latest version..."
  VERSION=$(get_latest_release)
  if [ -z "$VERSION" ]; then
    echo "Could not find latest release."
    exit 1
  fi
  echo "Latest version is $VERSION"
  
  download_release "$VERSION"
  
  echo "$BINARY installed successfully!"
}

main
