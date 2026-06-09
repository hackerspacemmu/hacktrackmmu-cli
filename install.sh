#!/usr/bin/env bash

set -euo pipefail

# GitHub Repository Info
OWNER="hackerspacemmu"
REPO="hacktrackmmu-cli"

# Global temp directory for cleanup trap
TMP_DIR=""

cleanup() {
  if [ -n "${TMP_DIR:-}" ] && [ -d "${TMP_DIR}" ]; then
    rm -rf "${TMP_DIR}"
  fi
}

# Setup color outputs
setup_colors() {
  if [ -t 2 ] && [ -z "${NO_COLOR-}" ] && [ "${TERM-}" != "dumb" ]; then
    BOLD="$(tput bold)"
    GREEN="$(tput setaf 2)"
    RED="$(tput setaf 1)"
    BLUE="$(tput setaf 4)"
    RESET="$(tput sgr0)"
  else
    BOLD=""
    GREEN=""
    RED=""
    BLUE=""
    RESET=""
  fi
}

info() {
  echo -e "${BLUE}info:${RESET} $*" >&2
}

success() {
  echo -e "${GREEN}success:${RESET} ${BOLD}$*${RESET}" >&2
}

error() {
  echo -e "${RED}error:${RESET} $*" >&2
  exit 1
}

# Detect OS
detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "${os}" in
    darwin) echo "darwin" ;;
    linux) echo "linux" ;;
    *) error "Unsupported operating system: ${os}" ;;
  esac
}

# Detect Architecture
detect_arch() {
  local arch
  arch="$(uname -m)"
  case "${arch}" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) error "Unsupported architecture: ${arch}" ;;
  esac
}

# Fetch the latest release version or use custom VERSION
get_version() {
  if [ -n "${VERSION:-}" ]; then
    echo "${VERSION}"
    return
  fi

  local latest_url="https://github.com/${OWNER}/${REPO}/releases/latest"
  # Fetch the final redirect URL to get the latest tag name
  local tag
  tag=$(curl -sL -o /dev/null -w "%{url_effective}" "${latest_url}" | rev | cut -d'/' -f1 | rev)
  
  if [ -z "${tag}" ] || [ "${tag}" = "latest" ]; then
    # Fallback to API if redirect check failed
    tag=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  fi

  if [ -z "${tag}" ]; then
    error "Could not retrieve the latest version. Please specify VERSION env var (e.g. VERSION=v1.0.0)"
  fi

  echo "${tag}"
}

main() {
  setup_colors

  info "Detecting system environment..."
  local os
  os=$(detect_os)
  local arch
  arch=$(detect_arch)

  info "Fetching version details..."
  local tag
  tag=$(get_version)
  # Strip leading 'v' for release filenames, but keep tag intact for URL
  local version="${tag#v}"

  info "Target: ${os}/${arch} (Version: ${tag})"

  # Define file names
  local archive_name="${REPO}_${version}_${os}_${arch}.tar.gz"
  local download_url="https://github.com/${OWNER}/${REPO}/releases/download/${tag}/${archive_name}"

  # Create a secure temporary directory
  TMP_DIR=$(mktemp -d)
  trap 'cleanup' EXIT

  info "Downloading ${download_url}..."
  if ! curl -fsSL "$download_url" -o "${TMP_DIR}/${archive_name}"; then
    error "Failed to download release archive. Make sure version ${tag} is published."
  fi

  info "Extracting archive..."
  tar -xzf "${TMP_DIR}/${archive_name}" -C "${TMP_DIR}"

  # Determine destination directory
  local dest_dir="/usr/local/bin"
  if [ ! -w "${dest_dir}" ]; then
    # If /usr/local/bin is not writable, check if ~/.local/bin exists or fall back to it
    if [ -d "${HOME}/.local/bin" ]; then
      dest_dir="${HOME}/.local/bin"
    else
      info "Requires root permissions to install to ${dest_dir}."
      # Prompt/use sudo if it's available
      if command -v sudo >/dev/null 2>&1; then
        dest_dir="/usr/local/bin"
      else
        error "Cannot write to ${dest_dir} and sudo is not available."
      fi
    fi
  fi

  info "Installing ${REPO} to ${dest_dir}..."
  if [ -w "${dest_dir}" ]; then
    mv "${TMP_DIR}/${REPO}" "${dest_dir}/${REPO}"
    chmod +x "${dest_dir}/${REPO}"
  else
    sudo mv "${TMP_DIR}/${REPO}" "${dest_dir}/${REPO}"
    sudo chmod +x "${dest_dir}/${REPO}"
  fi

  success "Installed ${REPO} successfully to ${dest_dir}/${REPO}!"
  
  # Check if installed command is in path
  if ! command -v "${REPO}" >/dev/null 2>&1; then
    info "Note: ${dest_dir} is not in your PATH. You may need to add it to your shell configuration."
  fi
}

main "$@"
