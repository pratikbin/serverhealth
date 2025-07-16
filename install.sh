#!/bin/bash

# Universal installation script for ServerHealth
# This script detects the platform and installs the appropriate package

set -e

APP_NAME="serverhealth"
GITHUB_REPO="kailashvele/serverhealth"  # Replace with your GitHub username
LATEST_VERSION=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Unicode symbols
CHECKMARK="\033[0;32m\xE2\x9C\x93\033[0m"
CROSS="\033[0;31m\xE2\x9C\x97\033[0m"
INFO="\033[0;34m\xE2\x84\xB9\033[0m"
WARNING="\033[0;33m\xE2\x9A\xA0\033[0m"

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    ServerHealth Installer                    ║"
    echo "║                                                              ║"
    echo "║  A comprehensive server monitoring tool with Slack           ║"
    echo "║  integration for real-time alerts and notifications.         ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_status() {
    echo -e "${INFO} ${BLUE}$1${NC}"
}

print_success() {
    echo -e "${CHECKMARK} ${GREEN}$1${NC}"
}

print_warning() {
    echo -e "${WARNING} ${YELLOW}$1${NC}"
}

print_error() {
    echo -e "${CROSS} ${RED}$1${NC}"
}

# Detect operating system and architecture
detect_platform() {
    print_status "Detecting platform..."

    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    # Normalize architecture names
    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        i386|i686)
            ARCH="386"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    # Normalize OS names
    case $OS in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ARCH="amd64"  # Default to amd64 for Windows
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    print_success "Detected platform: $OS/$ARCH"

    # Set binary extension for Windows
    BINARY_EXT=""
    if [ "$OS" = "windows" ]; then
        BINARY_EXT=".exe"
    fi
}

# Check if running as root (for system-wide installation)
check_root() {
    if [ "$EUID" -eq 0 ]; then
        INSTALL_DIR="/usr/local/bin"
        SYSTEM_INSTALL=true
        print_status "Running as root - will install system-wide"
    else
        INSTALL_DIR="$HOME/.local/bin"
        SYSTEM_INSTALL=false
        print_warning "Not running as root - will install to user directory"
        mkdir -p "$INSTALL_DIR"
    fi
}

# Get latest version from GitHub
get_latest_version() {
    print_status "Fetching latest version..."

    if command -v curl >/dev/null 2>&1; then
        LATEST_VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    elif command -v wget >/dev/null 2>&1; then
        LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [ -z "$LATEST_VERSION" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi

    print_success "Latest version: $LATEST_VERSION"
}

# Download and install binary
download_and_install() {
    print_status "Downloading and installing..."

    # Construct download URL - handle both regular versions and dev versions
    PACKAGE_NAME="${APP_NAME}-${LATEST_VERSION}-${OS}-${ARCH}"

    # For the download URL, we need to add 'v' prefix if it's not a dev version
    if [[ "$LATEST_VERSION" == dev-* ]]; then
        TAG_NAME="v${LATEST_VERSION}"
    else
        TAG_NAME="v${LATEST_VERSION}"
    fi

    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/${TAG_NAME}/${PACKAGE_NAME}.zip"
        ARCHIVE_EXT="zip"
    else
        DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/${TAG_NAME}/${PACKAGE_NAME}.tar.gz"
        ARCHIVE_EXT="tar.gz"
    fi

    print_status "Download URL: $DOWNLOAD_URL"

    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"

    # Download the package
    print_status "Downloading package..."
    if command -v curl >/dev/null 2>&1; then
        curl -LO "$DOWNLOAD_URL" || {
            print_error "Failed to download package"
            print_error "URL: $DOWNLOAD_URL"
            cleanup_and_exit 1
        }
    elif command -v wget >/dev/null 2>&1; then
        wget "$DOWNLOAD_URL" || {
            print_error "Failed to download package"
            print_error "URL: $DOWNLOAD_URL"
            cleanup_and_exit 1
        }
    fi

    # Extract the package
    print_status "Extracting package..."
    if [ "$ARCHIVE_EXT" = "zip" ]; then
        unzip "${PACKAGE_NAME}.${ARCHIVE_EXT}" || {
            print_error "Failed to extract package"
            cleanup_and_exit 1
        }
    else
        tar -xzf "${PACKAGE_NAME}.${ARCHIVE_EXT}" || {
            print_error "Failed to extract package"
            cleanup_and_exit 1
        }
    fi

    # Debug: show what was extracted
    print_status "Extracted contents:"
    ls -la

    # Install binary - the extracted directory should contain the binary
    BINARY_PATH=""
    if [ -f "${PACKAGE_NAME}/${APP_NAME}${BINARY_EXT}" ]; then
        BINARY_PATH="${PACKAGE_NAME}/${APP_NAME}${BINARY_EXT}"
    elif [ -f "${APP_NAME}${BINARY_EXT}" ]; then
        BINARY_PATH="${APP_NAME}${BINARY_EXT}"
    else
        print_error "Binary not found in expected locations"
        print_error "Contents of extracted archive:"
        find . -name "*${APP_NAME}*" -type f
        cleanup_and_exit 1
    fi

    print_status "Installing binary from $BINARY_PATH to $INSTALL_DIR..."
    cp "$BINARY_PATH" "$INSTALL_DIR/" || {
        print_error "Failed to copy binary"
        cleanup_and_exit 1
    }

    chmod +x "${INSTALL_DIR}/${APP_NAME}${BINARY_EXT}" || {
        print_error "Failed to make binary executable"
        cleanup_and_exit 1
    }

    print_success "Binary installed successfully!"

    # Cleanup
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
}

# Add to PATH
setup_path() {
    if [ "$SYSTEM_INSTALL" = true ]; then
        print_success "Binary is available system-wide"
        return
    fi

    print_status "Setting up PATH..."

    # Check if directory is already in PATH
    if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        print_success "Directory already in PATH"
        return
    fi

    # Add to shell profile
    SHELL_NAME=$(basename "$SHELL")
    case $SHELL_NAME in
        bash)
            PROFILE_FILE="$HOME/.bashrc"
            ;;
        zsh)
            PROFILE_FILE="$HOME/.zshrc"
            ;;
        fish)
            PROFILE_FILE="$HOME/.config/fish/config.fish"
            ;;
        *)
            PROFILE_FILE="$HOME/.profile"
            ;;
    esac

    if [ -f "$PROFILE_FILE" ]; then
        echo "" >> "$PROFILE_FILE"
        echo "# Added by ServerHealth installer" >> "$PROFILE_FILE"
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$PROFILE_FILE"
        print_success "Added $INSTALL_DIR to PATH in $PROFILE_FILE"
        print_warning "Please restart your terminal or run: source $PROFILE_FILE"
    else
        print_warning "Could not find shell profile file"
        print_warning "Please add $INSTALL_DIR to your PATH manually"
    fi
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."

    # Try to find the binary
    if [ -f "${INSTALL_DIR}/${APP_NAME}${BINARY_EXT}" ]; then
        print_success "Binary installed at ${INSTALL_DIR}/${APP_NAME}${BINARY_EXT}"

        # Try to run it
        if "${INSTALL_DIR}/${APP_NAME}${BINARY_EXT}" --version >/dev/null 2>&1; then
            VERSION_OUTPUT=$("${INSTALL_DIR}/${APP_NAME}${BINARY_EXT}" --version 2>/dev/null || echo "installed")
            print_success "Installation verified: $VERSION_OUTPUT"
        else
            print_success "Binary installed but version check failed (this might be normal)"
        fi
        return 0
    else
        print_error "Installation verification failed - binary not found"
        return 1
    fi
}

# Show post-installation instructions
show_post_install_instructions() {
    echo ""
    echo -e "${GREEN}🎉 ServerHealth has been installed successfully!${NC}"
    echo ""

    # Show the correct command path
    if [ "$SYSTEM_INSTALL" = true ] || [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
        CMD_PREFIX="$APP_NAME"
    else
        CMD_PREFIX="${INSTALL_DIR}/${APP_NAME}${BINARY_EXT}"
    fi

    echo -e "${CYAN}Next steps:${NC}"
    echo -e "  1. Configure monitoring: ${YELLOW}${CMD_PREFIX} configure${NC}"
    echo -e "  2. Start monitoring: ${YELLOW}${CMD_PREFIX} start${NC}"

    if [ "$SYSTEM_INSTALL" = true ]; then
        echo -e "  3. Install as system service: ${YELLOW}sudo ${CMD_PREFIX} install${NC}"
        echo -e "  4. Check service status: ${YELLOW}systemctl status $APP_NAME${NC}"
    else
        echo -e "  3. Install as user service: ${YELLOW}${CMD_PREFIX} install${NC}"
    fi

    echo ""
    echo -e "${CYAN}Useful commands:${NC}"
    echo -e "  • Check status: ${YELLOW}${CMD_PREFIX} status${NC}"
    echo -e "  • View logs: ${YELLOW}${CMD_PREFIX} logs${NC}"
    echo -e "  • Stop monitoring: ${YELLOW}${CMD_PREFIX} stop${NC}"
    echo -e "  • Reconfigure: ${YELLOW}${CMD_PREFIX} configure${NC}"
    echo -e "  • Get help: ${YELLOW}${CMD_PREFIX} --help${NC}"

    echo ""
    echo -e "${BLUE}Documentation:${NC}"
    echo -e "  • GitHub: https://github.com/$GITHUB_REPO"
    echo -e "  • Issues: https://github.com/$GITHUB_REPO/issues"

    if [ "$SYSTEM_INSTALL" != true ] && [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo -e "${WARNING} ${YELLOW}Note: You may need to restart your terminal or run 'source ~/.bashrc' to use the command globally.${NC}"
        echo -e "${WARNING} ${YELLOW}Or use the full path: ${CMD_PREFIX}${NC}"
    fi

    echo ""
}

# Cleanup function
cleanup_and_exit() {
    local exit_code=${1:-0}
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
    exit $exit_code
}

# Check dependencies
check_dependencies() {
    print_status "Checking dependencies..."

    local missing_deps=()

    # Check for required tools
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        missing_deps+=("curl or wget")
    fi

    if [ "$OS" != "windows" ]; then
        if ! command -v tar >/dev/null 2>&1; then
            missing_deps+=("tar")
        fi
    else
        if ! command -v unzip >/dev/null 2>&1; then
            missing_deps+=("unzip")
        fi
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies:"
        for dep in "${missing_deps[@]}"; do
            echo -e "  ${CROSS} $dep"
        done
        echo ""
        print_status "Please install the missing dependencies and try again."
        exit 1
    fi

    print_success "All dependencies are available"
}

# Handle command line arguments
handle_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                LATEST_VERSION="$2"
                print_status "Using specified version: $LATEST_VERSION"
                shift 2
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            --force|-f)
                FORCE_INSTALL=true
                shift
                ;;
            --user)
                FORCE_USER_INSTALL=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    echo "ServerHealth Installation Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --version VERSION    Install specific version"
    echo "  --force, -f          Force installation even if already installed"
    echo "  --user               Force user installation (non-root)"
    echo "  --help, -h           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Install latest version"
    echo "  $0 --version 1.2.0          # Install specific version"
    echo "  $0 --user                   # Install to user directory"
    echo ""
}

# Check if already installed
check_existing_installation() {
    if [ "$FORCE_INSTALL" = true ]; then
        return
    fi

    if command -v "$APP_NAME" >/dev/null 2>&1; then
        print_warning "ServerHealth is already installed"

        echo ""
        echo "Do you want to:"
        echo "  1) Update to latest version"
        echo "  2) Reinstall current version"
        echo "  3) Cancel installation"
        echo ""
        read -p "Please choose [1-3]: " choice

        case $choice in
            1|2)
                print_status "Proceeding with installation..."
                ;;
            3|*)
                print_status "Installation cancelled"
                exit 0
                ;;
        esac
    fi
}

# Set trap for cleanup
trap 'cleanup_and_exit 1' INT TERM

# Main execution
main() {
    print_header

    handle_arguments "$@"
    check_dependencies
    check_existing_installation
    detect_platform

    # Force user install if requested
    if [ "$FORCE_USER_INSTALL" = true ]; then
        INSTALL_DIR="$HOME/.local/bin"
        SYSTEM_INSTALL=false
        mkdir -p "$INSTALL_DIR"
        print_status "Forcing user installation"
    else
        check_root
    fi

    if [ -z "$LATEST_VERSION" ]; then
        get_latest_version
    fi

    download_and_install
    setup_path

    if verify_installation; then
        show_post_install_instructions
    else
        print_error "Installation failed"
        exit 1
    fi
}

# Run main function
main "$@"
