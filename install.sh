#!/usr/bin/env bash
#
# Zenlayer Cloud CLI (zeno) installer
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash -s -- --version v0.1.0
#   curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | INSTALL_DIR=/opt/bin bash

set -euo pipefail

REPO="zenlayer/zenlayercloud-cli"
BINARY="zeno"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# --- helpers ---

info()  { printf "\033[1;32m==>\033[0m %s\n" "$*"; }
warn()  { printf "\033[1;33mWARN:\033[0m %s\n" "$*" >&2; }
error() { printf "\033[1;31mERROR:\033[0m %s\n" "$*" >&2; exit 1; }

need_cmd() {
    if ! command -v "$1" &>/dev/null; then
        error "Required command '$1' not found. Please install it first."
    fi
}

# --- detect platform ---

detect_os() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux)  echo "linux" ;;
        Darwin) echo "darwin" ;;
        *)      error "Unsupported OS: $os" ;;
    esac
}

detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)   echo "arm64" ;;
        *)               error "Unsupported architecture: $arch" ;;
    esac
}

# --- main ---

main() {
    need_cmd curl
    need_cmd tar

    local version=""

    # parse args
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --version|-v) version="$2"; shift 2 ;;
            *)            error "Unknown option: $1" ;;
        esac
    done

    local os arch
    os="$(detect_os)"
    arch="$(detect_arch)"

    # For macOS, GoReleaser produces a universal binary with arch "all"
    if [[ "$os" == "darwin" ]]; then
        arch="all"
    fi

    info "Detected platform: ${os}/${arch}"

    # Resolve version
    if [[ -z "$version" ]]; then
        info "Fetching latest release..."
        version="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep -o '"tag_name": *"[^"]*"' | head -1 | grep -o '"v[^"]*"' | tr -d '"')"
        if [[ -z "$version" ]]; then
            error "Failed to determine latest version. Specify one with --version."
        fi
    fi

    # Strip leading 'v' for the archive filename (goreleaser uses version without v)
    local ver_num="${version#v}"

    info "Installing ${BINARY} ${version}..."

    # Build download URL
    # Archive naming from goreleaser: zeno_<VERSION>_<OS>_<ARCH>.tar.gz
    local archive="${BINARY}_${ver_num}_${os}_${arch}.tar.gz"
    local url="https://github.com/${REPO}/releases/download/${version}/${archive}"

    # Create temp directory
    local tmpdir
    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    info "Downloading ${url}..."
    if ! curl -fSL --progress-bar -o "${tmpdir}/${archive}" "$url"; then
        error "Download failed. Check that version '${version}' exists and has a release asset for ${os}/${arch}."
    fi

    # Verify checksum if available
    local checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"
    if curl -fsSL -o "${tmpdir}/checksums.txt" "$checksums_url" 2>/dev/null; then
        info "Verifying checksum..."
        local expected actual
        expected="$(grep "${archive}" "${tmpdir}/checksums.txt" | awk '{print $1}')"
        if [[ -n "$expected" ]]; then
            if command -v sha256sum &>/dev/null; then
                actual="$(sha256sum "${tmpdir}/${archive}" | awk '{print $1}')"
            elif command -v shasum &>/dev/null; then
                actual="$(shasum -a 256 "${tmpdir}/${archive}" | awk '{print $1}')"
            else
                warn "No sha256sum/shasum found, skipping checksum verification."
                actual="$expected"
            fi
            if [[ "$expected" != "$actual" ]]; then
                error "Checksum mismatch!\n  Expected: ${expected}\n  Actual:   ${actual}"
            fi
            info "Checksum verified."
        fi
    fi

    # Extract
    info "Extracting..."
    tar -xzf "${tmpdir}/${archive}" -C "${tmpdir}"

    if [[ ! -f "${tmpdir}/${BINARY}" ]]; then
        error "Binary '${BINARY}' not found in archive."
    fi

    chmod +x "${tmpdir}/${BINARY}"

    # Install
    if [[ -w "$INSTALL_DIR" ]]; then
        mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        info "Elevating privileges to install to ${INSTALL_DIR}..."
        sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    info "Installed ${BINARY} ${version} to ${INSTALL_DIR}/${BINARY}"

    # Verify
    if command -v "$BINARY" &>/dev/null; then
        info "Run '${BINARY} --help' to get started."
    else
        warn "${INSTALL_DIR} is not in your PATH. Add it with:"
        warn "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

main "$@"
