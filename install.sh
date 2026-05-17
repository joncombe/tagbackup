#!/bin/sh
set -e

REPO="joncombe/tagbackup"
BINARY="tagbackup"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
  Linux*)  OS=linux ;;
  Darwin*) OS=darwin ;;
  *)
    echo "error: unsupported operating system: ${OS}" >&2
    echo "Download a binary manually from https://github.com/${REPO}/releases" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64)        ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *)
    echo "error: unsupported architecture: ${ARCH}" >&2
    echo "Download a binary manually from https://github.com/${REPO}/releases" >&2
    exit 1
    ;;
esac

# Resolve version (use VERSION env var or fetch latest from GitHub API)
if [ -z "${VERSION}" ]; then
  if command -v curl >/dev/null 2>&1; then
    VERSION="$(curl -sfL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
  elif command -v wget >/dev/null 2>&1; then
    VERSION="$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
  else
    echo "error: curl or wget is required" >&2
    exit 1
  fi
  if [ -z "${VERSION}" ]; then
    echo "error: could not determine latest version from GitHub API" >&2
    exit 1
  fi
fi

# Construct download URL (VERSION may have a leading "v" from the git tag, but
# GoReleaser strips it from archive filenames via {{ .Version }})
VERSION_BARE="${VERSION#v}"
FILENAME="${BINARY}_${VERSION_BARE}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

# Determine install directory
if [ -z "${INSTALL_DIR}" ]; then
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "${INSTALL_DIR}"
  fi
fi

if [ ! -d "${INSTALL_DIR}" ]; then
  echo "error: install directory does not exist: ${INSTALL_DIR}" >&2
  exit 1
fi

if [ ! -w "${INSTALL_DIR}" ]; then
  echo "error: install directory is not writable: ${INSTALL_DIR}" >&2
  echo "Try: sudo INSTALL_DIR=${INSTALL_DIR} sh install.sh" >&2
  exit 1
fi

# Download and install
TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Downloading ${BINARY} ${VERSION} (${OS}/${ARCH})..."

if command -v curl >/dev/null 2>&1; then
  curl -fL --progress-bar "${URL}" -o "${TMP}/${FILENAME}" || {
    echo "error: download failed: ${URL}" >&2
    exit 1
  }
else
  wget -O "${TMP}/${FILENAME}" "${URL}" || {
    echo "error: download failed: ${URL}" >&2
    exit 1
  }
fi

if [ ! -s "${TMP}/${FILENAME}" ]; then
  echo "error: downloaded file is empty: ${URL}" >&2
  exit 1
fi

tar -xzf "${TMP}/${FILENAME}" -C "${TMP}" || {
  echo "error: failed to extract archive" >&2
  exit 1
}
install -m 755 "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "${BINARY} ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"

# Warn if the install directory is not in PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "NOTE: ${INSTALL_DIR} is not in your PATH."
    echo "Add it by running:"
    echo ""
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
    echo ""
    echo "Or for bash:"
    echo ""
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
    ;;
esac
