#!/usr/bin/env bash
# Install script for akv CLI tool
# Supports Linux and macOS with automatic platform detection
# Usage: ./install.sh [VERSION]
# Environment variables:
#   AKV_INSTALL_DIR - Installation directory (default: ~/.local/bin)
#   GITHUB_TOKEN    - GitHub token for API authentication (optional)

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="voidbear-io/akv"
BINARY_NAME="akv"

# Get default install directory
get_default_install_dir() {
    if [[ -n "${AKV_INSTALL_DIR:-}" ]]; then
        echo "$AKV_INSTALL_DIR"
    else
        echo "$HOME/.local/bin"
    fi
}

# Detect OS
detect_os() {
    local os
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux*)     echo "linux" ;;
        darwin*)    echo "darwin" ;;
        *)          echo "unknown" ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        *)              echo "unknown" ;;
    esac
}

# Get latest version from GitHub
get_latest_version() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        if [[ -n "${GITHUB_TOKEN:-}" ]]; then
            curl -s -H "Authorization: token $GITHUB_TOKEN" "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
        else
            curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
        fi
    elif command -v wget >/dev/null 2>&1; then
        if [[ -n "${GITHUB_TOKEN:-}" ]]; then
            wget -q --header="Authorization: token $GITHUB_TOKEN" -O - "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
        else
            wget -q -O - "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
        fi
    else
        echo "Error: Neither curl nor wget is installed" >&2
        exit 1
    fi
}

# Download file
download_file() {
    local url="$1"
    local output="$2"
    
    echo "Downloading from $url..."
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL -o "$output" "$url"; then
            echo -e "${RED}Error: Failed to download from $url${NC}" >&2
            return 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -q -O "$output" "$url"; then
            echo -e "${RED}Error: Failed to download from $url${NC}" >&2
            return 1
        fi
    else
        echo -e "${RED}Error: Neither curl nor wget is installed${NC}" >&2
        return 1
    fi
    
    return 0
}

# Main installation function
main() {
    local version="${1:-}"
    local install_dir
    local os
    local arch
    local download_url
    local temp_dir
    local archive_name
    local binary_path
    
    # Detect platform
    os=$(detect_os)
    arch=$(detect_arch)
    
    if [[ "$os" == "unknown" ]]; then
        echo -e "${RED}Error: Unsupported operating system${NC}" >&2
        exit 1
    fi
    
    if [[ "$arch" == "unknown" ]]; then
        echo -e "${RED}Error: Unsupported architecture: $(uname -m)${NC}" >&2
        exit 1
    fi
    
    echo "Detected platform: ${os}/${arch}"
    
    # Get version
    if [[ -z "$version" ]]; then
        echo "Detecting latest version..."
        version=$(get_latest_version)
        if [[ -z "$version" ]]; then
            echo -e "${RED}Error: Could not detect latest version${NC}" >&2
            exit 1
        fi
        echo "Latest version: $version"
    fi
    
    # Remove 'v' prefix if present for URL construction
    local version_for_url="${version#v}"
    
    # Set install directory
    install_dir=$(get_default_install_dir)
    echo "Install directory: $install_dir"
    
    # Create install directory if it doesn't exist
    if [[ ! -d "$install_dir" ]]; then
        echo "Creating install directory..."
        mkdir -p "$install_dir"
    fi
    
    # Construct download URL
    archive_name="${BINARY_NAME}-${os}-${arch}-v${version_for_url}.tar.gz"
    download_url="https://github.com/${REPO}/releases/download/v${version_for_url}/${archive_name}"
    
    echo "Downloading ${archive_name}..."
    
    # Create temp directory
    temp_dir=$(mktemp -d)
    trap "rm -rf '$temp_dir'" EXIT
    
    # Download archive
    local archive_path="${temp_dir}/${archive_name}"
    if ! download_file "$download_url" "$archive_path"; then
        # Try with 'v' prefix
        archive_name="${BINARY_NAME}-${os}-${arch}-${version}.tar.gz"
        download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
        echo "Retrying with alternate URL format..."
        if ! download_file "$download_url" "$archive_path"; then
            echo -e "${RED}Error: Failed to download release archive${NC}" >&2
            exit 1
        fi
    fi
    
    echo -e "${GREEN}Download complete${NC}"
    
    # Extract archive
    echo "Extracting archive..."
    cd "$temp_dir"
    if ! tar -xzf "$archive_name"; then
        echo -e "${RED}Error: Failed to extract archive${NC}" >&2
        exit 1
    fi
    
    # Find binary in extracted directory
    binary_path=$(find "$temp_dir" -name "$BINARY_NAME" -type f | head -1)
    if [[ -z "$binary_path" ]]; then
        echo -e "${RED}Error: Could not find binary in archive${NC}" >&2
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$binary_path"
    
    # Install binary
    local install_path="${install_dir}/${BINARY_NAME}"
    echo "Installing to ${install_path}..."
    
    # Check if we need sudo for system directories
    if [[ "$install_dir" == /usr/* || "$install_dir" == /opt/* ]]; then
        if [[ -w "$install_dir" ]]; then
            cp "$binary_path" "$install_path"
        else
            echo "Elevated permissions required for ${install_dir}..."
            sudo cp "$binary_path" "$install_path"
        fi
    else
        cp "$binary_path" "$install_path"
    fi
    
    # Verify installation
    if [[ -x "$install_path" ]]; then
        echo -e "${GREEN}✓ ${BINARY_NAME} ${version} installed successfully!${NC}"
        
        # Check if install directory is in PATH
        if [[ ":$PATH:" != *":${install_dir}:"* ]]; then
            echo ""
            echo -e "${YELLOW}Warning: ${install_dir} is not in your PATH${NC}"
            echo "Add the following to your shell profile:"
            echo "  export PATH=\"\$PATH:${install_dir}\""
        fi
        
        # Show version
        echo ""
        echo "Installed version:"
        "$install_path" --version 2>/dev/null || echo "(version command not available)"
    else
        echo -e "${RED}Error: Installation failed${NC}" >&2
        exit 1
    fi
}

# Run main function
main "$@"
