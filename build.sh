#!/bin/bash

# Build script for serverhealth
# This script builds the application for multiple platforms

set -e

# Application name and version
APP_NAME="serverhealth"
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_DIR="build"
DIST_DIR="dist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean previous builds
clean() {
    print_status "Cleaning previous builds..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$BUILD_DIR" "$DIST_DIR"
}

# Get dependencies
get_deps() {
    print_status "Getting dependencies..."
    go mod tidy
    go mod download
}

# Build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local ext=$3

    local output_name="${APP_NAME}"
    if [ "$ext" != "" ]; then
        output_name="${output_name}.${ext}"
    fi

    local output_path="${BUILD_DIR}/${APP_NAME}-${os}-${arch}/${output_name}"

    print_status "Building for ${os}/${arch}..."

    mkdir -p "$(dirname "$output_path")"

    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -ldflags="-w -s -X main.version=${VERSION}" \
        -o "$output_path" \
        .

    if [ $? -eq 0 ]; then
        print_success "Built ${os}/${arch} -> $output_path"

        # Create distribution package
        create_package "$os" "$arch" "$output_path"
    else
        print_error "Failed to build for ${os}/${arch}"
        exit 1
    fi
}

# Create distribution package
create_package() {
    local os=$1
    local arch=$2
    local binary_path=$3

    local package_dir="${DIST_DIR}/${APP_NAME}-${VERSION}-${os}-${arch}"
    local package_name="${APP_NAME}-${VERSION}-${os}-${arch}"

    mkdir -p "$package_dir"

    # Copy binary
    cp "$binary_path" "$package_dir/"

    # Copy documentation
    if [ -f "README.md" ]; then
        cp "README.md" "$package_dir/"
    fi

    if [ -f "LICENSE" ]; then
        cp "LICENSE" "$package_dir/"
    fi

    # Create install script for each platform
    create_install_script "$os" "$package_dir"

    # Create archive
    cd "$DIST_DIR"
    if [ "$os" = "windows" ]; then
        zip -r "${package_name}.zip" "$(basename "$package_dir")" > /dev/null
        print_success "Created ${package_name}.zip"
    else
        tar -czf "${package_name}.tar.gz" "$(basename "$package_dir")"
        print_success "Created ${package_name}.tar.gz"
    fi
    cd - > /dev/null
}

# Create platform-specific install script
create_install_script() {
    local os=$1
    local package_dir=$2

    case $os in
        "linux"|"darwin")
            cat > "$package_dir/install.sh" << 'EOF'
#!/bin/bash

set -e

APP_NAME="serverhealth"
INSTALL_DIR="/usr/local/bin"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_warning "Installing to user directory instead of system-wide"
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Copy binary
print_status "Installing $APP_NAME to $INSTALL_DIR..."
cp "./$APP_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$APP_NAME"

# Add to PATH if not already there
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    print_warning "Add $INSTALL_DIR to your PATH to use $APP_NAME from anywhere"
    echo "Add this line to your ~/.bashrc or ~/.zshrc:"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi

print_status "Installation complete!"
print_status "Run '$APP_NAME configure' to get started"
EOF
            chmod +x "$package_dir/install.sh"
            ;;
        "windows")
            cat > "$package_dir/install.bat" << 'EOF'
@echo off
setlocal

set APP_NAME=serverhealth
set INSTALL_DIR=%ProgramFiles%\%APP_NAME%

echo Installing %APP_NAME%...

REM Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Error: Administrator privileges required
    echo Please run this script as Administrator
    pause
    exit /b 1
)

REM Create install directory
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM Copy binary
copy "%APP_NAME%.exe" "%INSTALL_DIR%\"

REM Add to PATH
setx PATH "%PATH%;%INSTALL_DIR%" /M

echo Installation complete!
echo Run '%APP_NAME% configure' to get started
pause
EOF
            ;;
    esac
}

# Build for all platforms
build_all() {
    print_status "Building for all platforms..."

    # Linux
    build_platform "linux" "amd64" ""
    build_platform "linux" "arm64" ""
    build_platform "linux" "386" ""

    # macOS
    build_platform "darwin" "amd64" ""
    build_platform "darwin" "arm64" ""

    # Windows
    build_platform "windows" "amd64" "exe"
    build_platform "windows" "386" "exe"

    # FreeBSD
    build_platform "freebsd" "amd64" ""
}

# Generate checksums
generate_checksums() {
    print_status "Generating checksums..."
    cd "$DIST_DIR"

    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum *.tar.gz *.zip > checksums.txt 2>/dev/null || true
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 *.tar.gz *.zip > checksums.txt 2>/dev/null || true
    fi

    if [ -f "checksums.txt" ]; then
        print_success "Generated checksums.txt"
    fi

    cd - > /dev/null
}

# Main execution
main() {
    print_status "Starting build process for $APP_NAME v$VERSION"

    clean
    get_deps
    build_all
    generate_checksums

    print_success "Build process completed!"
    print_status "Distribution packages created in $DIST_DIR/"

    # List created packages
    echo ""
    echo "Created packages:"
    ls -la "$DIST_DIR/" | grep -E '\.(tar\.gz|zip)$' || true
}

# Help function
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build     Build for all platforms (default)"
    echo "  clean     Clean build artifacts"
    echo "  deps      Get dependencies"
    echo "  help      Show this help"
    echo ""
    echo "Environment variables:"
    echo "  VERSION   Set version number (default: git describe)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Build for all platforms"
    echo "  VERSION=1.1.0 $0     # Build with custom version"
    echo "  $0 clean             # Clean build artifacts"
}

# Handle command line arguments
case "${1:-build}" in
    "build")
        main
        ;;
    "clean")
        clean
        print_success "Cleaned build artifacts"
        ;;
    "deps")
        get_deps
        print_success "Dependencies updated"
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
