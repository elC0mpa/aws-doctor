#!/bin/sh
set -e

REPO="elC0mpa/aws-doctor"
BINARY_NAME="aws-doctor"

if [ -z "$INSTALL_DIR" ]; then
    if [ "$(id -u)" -eq 0 ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
    fi
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    printf "${GREEN}[INFO]${NC} %s\n" "$1" >&2
}

log_warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1" >&2
}

log_error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1" >&2
}

detect_os() {
    OS="$(uname -s)"
    case "$OS" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
}

detect_arch() {
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
}

get_latest_version() {
    curl -sS "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/'
}

download_and_verify() {
    VERSION="$1"
    OS="$2"
    ARCH="$3"
    TMP_DIR="$4"

    if [ "$OS" = "windows" ]; then
        EXT="zip"
    else
        EXT="tar.gz"
    fi

    FILENAME="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.${EXT}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
    CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    log_info "Downloading ${FILENAME}..."
    
    curl -sSfL "$DOWNLOAD_URL" -o "${TMP_DIR}/${FILENAME}"
    curl -sSfL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.txt"

    log_info "Verifying checksum..."
    cd "$TMP_DIR"

    EXPECTED_CHECKSUM=$(grep "${FILENAME}" checksums.txt | awk '{print $1}')
    if [ -z "$EXPECTED_CHECKSUM" ]; then
        log_error "Could not find checksum for ${FILENAME}"
        exit 1
    fi

    if command -v sha256sum > /dev/null 2>&1; then
        ACTUAL_CHECKSUM=$(sha256sum "${FILENAME}" | awk '{print $1}')
    elif command -v shasum > /dev/null 2>&1; then
        ACTUAL_CHECKSUM=$(shasum -a 256 "${FILENAME}" | awk '{print $1}')
    else
        log_warn "Neither sha256sum nor shasum found, skipping checksum verification"
        ACTUAL_CHECKSUM="$EXPECTED_CHECKSUM"
    fi

    if [ "$EXPECTED_CHECKSUM" != "$ACTUAL_CHECKSUM" ]; then
        log_error "Checksum verification failed!"
        log_error "Expected: $EXPECTED_CHECKSUM"
        log_error "Actual:   $ACTUAL_CHECKSUM"
        exit 1
    fi

    log_info "Checksum verified successfully"

    log_info "Extracting archive..."
    if [ "$EXT" = "zip" ]; then
        unzip -q "${FILENAME}"
    else
        tar -xzf "${FILENAME}"
    fi

    echo "${TMP_DIR}/${BINARY_NAME}"
}

install_binary() {
    BINARY_PATH="$1"
    TARGET_FILE="${INSTALL_DIR}/${BINARY_NAME}"

    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR" || {
            log_error "Failed to create installation directory: $INSTALL_DIR"
            exit 1
        }
    fi

    if [ -w "$INSTALL_DIR" ]; then
        if [ -f "$TARGET_FILE" ] && [ ! -w "$TARGET_FILE" ]; then
             log_error "Target file $TARGET_FILE exists and is not writable."
             log_error "It might be owned by root. Try running: sudo rm $TARGET_FILE"
             exit 1
        fi
        install -m 755 "$BINARY_PATH" "$TARGET_FILE"
    else
        # If the directory is not writable and is inside $HOME, fail instead of using sudo
        case "$INSTALL_DIR" in
            "$HOME"*)
                log_error "Installation directory $INSTALL_DIR is not writable."
                log_error "Please check permissions."
                exit 1
                ;;
            *)
                log_info "Installing to ${INSTALL_DIR} requires elevated permissions..."
                sudo install -m 755 "$BINARY_PATH" "$TARGET_FILE"
                ;;
        esac
    fi
}

main() {
    log_info "AWS Doctor Installer"
    log_info "===================="

    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    OS=$(detect_os)
    ARCH=$(detect_arch)

    log_info "Detected OS: $OS"
    log_info "Detected architecture: $ARCH"

    VERSION="${1:-$(get_latest_version)}"
    if [ -z "$VERSION" ]; then
        log_error "Could not determine version to install"
        exit 1
    fi

    log_info "Installing version: $VERSION"

    BINARY_PATH=$(download_and_verify "$VERSION" "$OS" "$ARCH" "$TMP_DIR")

    install_binary "$BINARY_PATH"

    log_info "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
    
    case ":$PATH:" in
        *":$INSTALL_DIR:"*) ;;
        *) log_warn "${INSTALL_DIR} is not in your PATH. Please add it to use ${BINARY_NAME}." ;;
    esac

    log_info ""
    log_info "Run 'aws-doctor --help' to get started"
}

main "$@"
