#!/bin/bash

# RPM package creation script
# This script creates .rpm packages for Red Hat/CentOS/Fedora systems

set -e

APP_NAME="serverhealth"
VERSION=${VERSION:-"1.0.0"}
RELEASE=${RELEASE:-"1"}
ARCHITECTURE=${ARCHITECTURE:-"x86_64"}
MAINTAINER="Your Name <your.email@example.com>"
DESCRIPTION="Server health monitoring tool with Slack notifications"

BUILD_DIR="build/rpm"
SPEC_FILE="${BUILD_DIR}/SPECS/${APP_NAME}.spec"

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

# Create RPM build structure
create_build_structure() {
    print_status "Creating RPM build structure..."

    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}

    # Copy source files
    mkdir -p "${BUILD_DIR}/SOURCES/${APP_NAME}-${VERSION}"
    cp "build/${APP_NAME}-linux-amd64/${APP_NAME}" "${BUILD_DIR}/SOURCES/${APP_NAME}-${VERSION}/"

    # Create source tarball
    cd "${BUILD_DIR}/SOURCES"
    tar -czf "${APP_NAME}-${VERSION}.tar.gz" "${APP_NAME}-${VERSION}/"
    rm -rf "${APP_NAME}-${VERSION}"
    cd - > /dev/null
}

# Create RPM spec file
create_spec_file() {
    print_status "Creating RPM spec file..."

    cat > "$SPEC_FILE" << EOF
Name:           ${APP_NAME}
Version:        ${VERSION}
Release:        ${RELEASE}%{?dist}
Summary:        ${DESCRIPTION}

License:        MIT
URL:            https://github.com/yourusername/%{name}
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  systemd-rpm-macros
Requires:       systemd
BuildArch:      ${ARCHITECTURE}

%description
A comprehensive server monitoring tool that tracks disk usage, CPU usage,
and memory usage, sending notifications to Slack when thresholds are exceeded.

Features:
- Cross-platform support (Linux, macOS, Windows)
- Interactive CLI configuration
- Slack integration
- Systemd service support
- Configurable thresholds and intervals

%prep
%setup -q

%build
# Binary is pre-built

%install
rm -rf %{buildroot}

# Install binary
mkdir -p %{buildroot}%{_bindir}
install -m 0755 %{name} %{buildroot}%{_bindir}/%{name}

# Install systemd service file
mkdir -p %{buildroot}%{_unitdir}
cat > %{buildroot}%{_unitdir}/%{name}.service << 'SERVICE_EOF'
[Unit]
Description=ServerHealth - Server Health Monitoring Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=5
User=nobody
Group=nobody
ExecStart=%{_bindir}/%{name} daemon
Environment=PATH=/usr/local/bin:/usr/bin:/bin
StandardOutput=journal
StandardError=journal
SyslogIdentifier=%{name}

[Install]
WantedBy=multi-user.target
SERVICE_EOF

# Install documentation
mkdir -p %{buildroot}%{_docdir}/%{name}
cat > %{buildroot}%{_docdir}/%{name}/README << 'README_EOF'
ServerHealth
============

A comprehensive server monitoring tool with Slack integration.

Quick Start:
1. Configure: %{name} configure
2. Start: systemctl start %{name}
3. Enable: systemctl enable %{name}

Configuration is stored in ~/.config/%{name}/

For more information:
https://github.com/yourusername/%{name}
README_EOF

%files
%{_bindir}/%{name}
%{_unitdir}/%{name}.service
%{_docdir}/%{name}/README

%post
%systemd_post %{name}.service
echo "ServerHealth installed successfully!"
echo "Run '%{name} configure' to set up monitoring"
echo "Then 'systemctl start %{name}' to start the service"

%preun
%systemd_preun %{name}.service

%postun
%systemd_postun_with_restart %{name}.service

%changelog
* $(date +'%a %b %d %Y') ${MAINTAINER} - ${VERSION}-${RELEASE}
- Initial RPM package release
- Cross-platform monitoring support
- Slack integration
- Systemd service support
EOF
}

# Build the RPM package
build_package() {
    print_status "Building RPM package..."

    rpmbuild --define "_topdir $(pwd)/${BUILD_DIR}" \
             --define "_binary_payload w2.xzdio" \
             -ba "$SPEC_FILE"

    # Move to dist directory
    mkdir -p dist
    find "${BUILD_DIR}/RPMS" -name "*.rpm" -exec cp {} dist/ \;
    find "${BUILD_DIR}/SRPMS" -name "*.rpm" -exec cp {} dist/ \;

    print_success "RPM packages created in dist/"
    ls -la dist/*.rpm
}

# Main execution
main() {
    if [ ! -f "build/${APP_NAME}-linux-amd64/${APP_NAME}" ]; then
        print_error "Binary not found. Please run './build.sh' first."
        exit 1
    fi

    create_build_structure
    create_spec_file
    build_package

    print_success "RPM package created successfully!"
    print_status "Install with: sudo rpm -ivh dist/${APP_NAME}-${VERSION}-${RELEASE}.*.rpm"
    print_status "Or with: sudo dnf install dist/${APP_NAME}-${VERSION}-${RELEASE}.*.rpm"
}

# Check if rpmbuild is available
if ! command -v rpmbuild >/dev/null 2>&1; then
    print_error "rpmbuild not found. Please install rpm-build package."
    print_status "On CentOS/RHEL: sudo yum install rpm-build"
    print_status "On Fedora: sudo dnf install rpm-build"
    exit 1
fi

main
