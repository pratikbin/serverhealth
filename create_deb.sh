#!/bin/bash

# Debian package creation script
# This script creates .deb packages for Ubuntu/Debian systems

set -e

APP_NAME="serverhealth"
VERSION=${VERSION:-"1.0.0"}
ARCHITECTURE=${ARCHITECTURE:-"amd64"}
MAINTAINER="Your Name <your.email@example.com>"
DESCRIPTION="Server health monitoring tool with Slack notifications"

BUILD_DIR="build/debian"
PACKAGE_DIR="${BUILD_DIR}/${APP_NAME}_${VERSION}_${ARCHITECTURE}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create package structure
create_package_structure() {
    print_status "Creating package structure..."

    rm -rf "$BUILD_DIR"
    mkdir -p "$PACKAGE_DIR"/{DEBIAN,usr/local/bin,etc/systemd/system,usr/share/doc/${APP_NAME}}

    # Copy binary
    cp "build/${APP_NAME}-linux-${ARCHITECTURE}/${APP_NAME}" "${PACKAGE_DIR}/usr/local/bin/"
    chmod +x "${PACKAGE_DIR}/usr/local/bin/${APP_NAME}"

    # Create systemd service file
    cat > "${PACKAGE_DIR}/etc/systemd/system/${APP_NAME}.service" << EOF
[Unit]
Description=ServerHealth - Server Health Monitoring Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=5
User=nobody
Group=nogroup
ExecStart=/usr/local/bin/${APP_NAME} daemon
Environment=PATH=/usr/local/bin:/usr/bin:/bin
StandardOutput=journal
StandardError=journal
SyslogIdentifier=${APP_NAME}

[Install]
WantedBy=multi-user.target
EOF
}

# Create control file
create_control_file() {
    print_status "Creating control file..."

    # Calculate installed size
    INSTALLED_SIZE=$(du -sk "$PACKAGE_DIR" | cut -f1)

    cat > "${PACKAGE_DIR}/DEBIAN/control" << EOF
Package: ${APP_NAME}
Version: ${VERSION}
Section: admin
Priority: optional
Architecture: ${ARCHITECTURE}
Installed-Size: ${INSTALLED_SIZE}
Maintainer: ${MAINTAINER}
Description: ${DESCRIPTION}
 A comprehensive server monitoring tool that tracks disk usage, CPU usage,
 and memory usage, sending notifications to Slack when thresholds are exceeded.
 .
 Features:
  - Cross-platform support (Linux, macOS, Windows)
  - Interactive CLI configuration
  - Slack integration
  - Systemd service support
  - Configurable thresholds and intervals
Homepage: https://github.com/yourusername/${APP_NAME}
EOF
}

# Create postinst script
create_postinst_script() {
    print_status "Creating postinst script..."

    cat > "${PACKAGE_DIR}/DEBIAN/postinst" << 'EOF'
#!/bin/bash

set -e

case "$1" in
    configure)
        # Reload systemd
        systemctl daemon-reload || true

        # Enable service but don't start it
        systemctl enable serverhealth.service || true

        echo "ServerHealth installed successfully!"
        echo "Run 'serverhealth configure' to set up monitoring"
        echo "Then 'systemctl start serverhealth' to start the service"
        ;;
esac

exit 0
EOF

    chmod +x "${PACKAGE_DIR}/DEBIAN/postinst"
}

# Create prerm script
create_prerm_script() {
    print_status "Creating prerm script..."

    cat > "${PACKAGE_DIR}/DEBIAN/prerm" << 'EOF'
#!/bin/bash

set -e

case "$1" in
    remove|deconfigure)
        # Stop and disable service
        systemctl stop serverhealth.service || true
        systemctl disable serverhealth.service || true
        ;;
esac

exit 0
EOF

    chmod +x "${PACKAGE_DIR}/DEBIAN/prerm"
}

# Create postrm script
create_postrm_script() {
    print_status "Creating postrm script..."

    cat > "${PACKAGE_DIR}/DEBIAN/postrm" << 'EOF'
#!/bin/bash

set -e

case "$1" in
    purge)
        # Remove configuration files
        rm -rf /home/*/.config/serverhealth || true
        ;;
    remove)
        # Reload systemd
        systemctl daemon-reload || true
        ;;
esac

exit 0
EOF

    chmod +x "${PACKAGE_DIR}/DEBIAN/postrm"
}

# Create documentation
create_documentation() {
    print_status "Creating documentation..."

    # Create README
    cat > "${PACKAGE_DIR}/usr/share/doc/${APP_NAME}/README.Debian" << EOF
ServerHealth for Debian/Ubuntu
==============================

This package provides the ServerHealth service for monitoring system
resources and sending notifications to Slack.

Quick Start:
1. Configure the monitor: serverhealth configure
2. Start the service: systemctl start serverhealth
3. Check status: systemctl status serverhealth

Configuration files are stored in ~/.config/serverhealth/

For more information, visit:
https://github.com/yourusername/serverhealth
EOF

    # Create changelog
    cat > "${PACKAGE_DIR}/usr/share/doc/${APP_NAME}/changelog.Debian" << EOF
serverhealth (${VERSION}) stable; urgency=medium

  * Initial Debian package release
  * Cross-platform monitoring support
  * Slack integration
  * Systemd service support

 -- ${MAINTAINER}  $(date -R)
EOF

    gzip "${PACKAGE_DIR}/usr/share/doc/${APP_NAME}/changelog.Debian"

    # Copy license if it exists
    if [ -f "LICENSE" ]; then
        cp "LICENSE" "${PACKAGE_DIR}/usr/share/doc/${APP_NAME}/copyright"
    fi
}

# Build the package
build_package() {
    print_status "Building Debian package..."

    cd "$BUILD_DIR"
    dpkg-deb --build "${APP_NAME}_${VERSION}_${ARCHITECTURE}"

    cd - > /dev/null

    # Move to dist directory
    mkdir -p dist
    mv "${BUILD_DIR}/${APP_NAME}_${VERSION}_${ARCHITECTURE}.deb" "dist/"

    print_success "Created dist/${APP_NAME}_${VERSION}_${ARCHITECTURE}.deb"
}

# Main execution
main() {
    if [ ! -f "build/${APP_NAME}-linux-${ARCHITECTURE}/${APP_NAME}" ]; then
        print_error "Binary not found. Please run './build.sh' first."
        exit 1
    fi

    create_package_structure
    create_control_file
    create_postinst_script
    create_prerm_script
    create_postrm_script
    create_documentation
    build_package

    print_success "Debian package created successfully!"
    print_status "Install with: sudo dpkg -i dist/${APP_NAME}_${VERSION}_${ARCHITECTURE}.deb"
}

# Check if dpkg-deb is available
if ! command -v dpkg-deb >/dev/null 2>&1; then
    print_error "dpkg-deb not found. Please install dpkg-dev package."
    exit 1
fi

main
