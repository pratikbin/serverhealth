#!/bin/bash

set -e

APP_NAME="serverhealth"
VERSION=${1:-"dev-$(date +%Y%m%d-%H%M%S)"}
BUILD_DIR="dist"
PKG_DIR="packages"

echo "🚀 Building $APP_NAME version $VERSION"

# Clean up
rm -rf "$BUILD_DIR" "$PKG_DIR"
mkdir -p "$BUILD_DIR" "$PKG_DIR"

# Build matrix - using separate arrays
platforms=(
    "linux-amd64"
    "linux-arm64"
    "linux-386"
    "windows-amd64"
    "windows-386"
    "darwin-amd64"
    "darwin-arm64"
    "freebsd-amd64"
)

echo "📦 Building binaries..."

for platform in "${platforms[@]}"; do
    # Split platform into GOOS and GOARCH
    IFS='-' read -r GOOS GOARCH <<< "$platform"

    BINARY_NAME="$APP_NAME"
    BINARY_EXT=""
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
        BINARY_EXT=".exe"
    fi

    echo "  Building for $GOOS/$GOARCH..."

    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=$VERSION -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') -s -w" \
        -o "$BUILD_DIR/${BINARY_NAME}-${platform}" .

    # Create archive with proper structure for install script
    cd "$BUILD_DIR"

    # Create directory for the archive (this is what install script expects)
    ARCHIVE_DIR="${APP_NAME}-${VERSION}-${platform}"
    mkdir -p "$ARCHIVE_DIR"

    # Copy binary with standard name (serverhealth or serverhealth.exe)
    cp "${BINARY_NAME}-${platform}" "$ARCHIVE_DIR/${APP_NAME}${BINARY_EXT}"

    # Create archive
    if [ "$GOOS" = "windows" ]; then
        zip -r "${APP_NAME}-${VERSION}-${platform}.zip" "$ARCHIVE_DIR"
    else
        tar -czf "${APP_NAME}-${VERSION}-${platform}.tar.gz" "$ARCHIVE_DIR"
    fi

    # Clean up temporary directory but keep the original binary for packaging
    rm -rf "$ARCHIVE_DIR"
    cd ..
done

echo "📋 Creating Linux packages..."

# Function to create DEB package
create_deb() {
    local arch=$1
    local deb_arch=$2

    echo "  Creating DEB package for $arch..."

    DEB_DIR="$PKG_DIR/deb-$arch"
    mkdir -p "$DEB_DIR/DEBIAN"
    mkdir -p "$DEB_DIR/usr/local/bin"
    mkdir -p "$DEB_DIR/etc/systemd/system"
    mkdir -p "$DEB_DIR/usr/share/doc/$APP_NAME"

    # Copy binary
    cp "$BUILD_DIR/$APP_NAME-linux-$arch" "$DEB_DIR/usr/local/bin/$APP_NAME"
    chmod +x "$DEB_DIR/usr/local/bin/$APP_NAME"

    # Create control file
    cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: $APP_NAME
Version: $VERSION
Architecture: $deb_arch
Maintainer: Your Name <your.email@example.com>
Description: Server health monitoring tool
 A cross-platform server monitoring tool that tracks CPU, memory, and disk usage
 with Slack notifications and service integration.
Homepage: https://github.com/yourusername/$APP_NAME
Section: admin
Priority: optional
Depends: systemd
EOF

    # Create postinst script
    cat > "$DEB_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Create user
if ! id serverhealth >/dev/null 2>&1; then
    useradd --system --no-create-home --shell /bin/false serverhealth
fi

# Create config directory
mkdir -p /etc/serverhealth
chown serverhealth:serverhealth /etc/serverhealth

# Reload systemd
systemctl daemon-reload

echo "Run 'sudo systemctl enable serverhealth' to start on boot"
echo "Run 'serverhealth configure' to set up monitoring"
EOF

    # Create prerm script
    cat > "$DEB_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e

if [ "$1" = "remove" ]; then
    systemctl stop serverhealth 2>/dev/null || true
    systemctl disable serverhealth 2>/dev/null || true
fi
EOF

    chmod +x "$DEB_DIR/DEBIAN/postinst" "$DEB_DIR/DEBIAN/prerm"

    # Create systemd service
    cat > "$DEB_DIR/etc/systemd/system/$APP_NAME.service" << EOF
[Unit]
Description=Server Health Monitor
After=network.target

[Service]
Type=simple
User=serverhealth
Group=serverhealth
ExecStart=/usr/local/bin/$APP_NAME start
Restart=always
RestartSec=30
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

    # Create documentation
    cat > "$DEB_DIR/usr/share/doc/$APP_NAME/README" << EOF
Server Health Monitor v$VERSION

To configure: serverhealth configure
To start: sudo systemctl start serverhealth
To enable on boot: sudo systemctl enable serverhealth

For more information, visit:
https://github.com/yourusername/$APP_NAME
EOF

    # Build DEB package
    dpkg-deb --build "$DEB_DIR" "$BUILD_DIR/${APP_NAME}_${VERSION}_${deb_arch}.deb"
}

# Function to create RPM package
create_rpm() {
    echo "  Creating RPM package..."

    # Setup RPM build environment
    RPM_ROOT="$PKG_DIR/rpm"
    mkdir -p "$RPM_ROOT"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

    # Create spec file
    cat > "$RPM_ROOT/SPECS/$APP_NAME.spec" << EOF
Name: $APP_NAME
Version: $VERSION
Release: 1%{?dist}
Summary: Server health monitoring tool
License: MIT
URL: https://github.com/yourusername/$APP_NAME
BuildArch: x86_64
Requires: systemd

%description
A cross-platform server monitoring tool that tracks CPU, memory, and disk usage
with Slack notifications and service integration.

%install
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/etc/systemd/system
mkdir -p %{buildroot}/usr/share/doc/$APP_NAME

cp %{_sourcedir}/$APP_NAME %{buildroot}/usr/local/bin/
cp %{_sourcedir}/$APP_NAME.service %{buildroot}/etc/systemd/system/
cp %{_sourcedir}/README %{buildroot}/usr/share/doc/$APP_NAME/

%files
/usr/local/bin/$APP_NAME
/etc/systemd/system/$APP_NAME.service
/usr/share/doc/$APP_NAME/README

%pre
getent passwd serverhealth >/dev/null || useradd --system --no-create-home --shell /bin/false serverhealth

%post
mkdir -p /etc/serverhealth
chown serverhealth:serverhealth /etc/serverhealth
systemctl daemon-reload
echo "Run 'sudo systemctl enable $APP_NAME' to start on boot"

%preun
if [ \$1 -eq 0 ]; then
    systemctl stop $APP_NAME 2>/dev/null || true
    systemctl disable $APP_NAME 2>/dev/null || true
fi

%postun
systemctl daemon-reload
EOF

    # Copy sources
    cp "$BUILD_DIR/$APP_NAME-linux-amd64" "$RPM_ROOT/SOURCES/$APP_NAME"

    # Create systemd service for RPM
    cat > "$RPM_ROOT/SOURCES/$APP_NAME.service" << EOF
[Unit]
Description=Server Health Monitor
After=network.target

[Service]
Type=simple
User=serverhealth
Group=serverhealth
ExecStart=/usr/local/bin/$APP_NAME start
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF

    # Create README
    cat > "$RPM_ROOT/SOURCES/README" << EOF
Server Health Monitor v$VERSION

To configure: $APP_NAME configure
To start: sudo systemctl start $APP_NAME
To enable on boot: sudo systemctl enable $APP_NAME
EOF

    # Build RPM (if rpmbuild is available)
    if command -v rpmbuild >/dev/null 2>&1; then
        rpmbuild --define "_topdir $(pwd)/$RPM_ROOT" -ba "$RPM_ROOT/SPECS/$APP_NAME.spec"
        cp "$RPM_ROOT/RPMS/x86_64"/*.rpm "$BUILD_DIR/"
    else
        echo "  ⚠️  rpmbuild not found, skipping RPM creation"
        echo "     Install rpm-build to create RPM packages"
    fi
}

# Create packages for Linux
if command -v dpkg-deb >/dev/null 2>&1; then
    create_deb "amd64" "amd64"
    create_deb "arm64" "arm64"
else
    echo "⚠️  dpkg-deb not found, skipping DEB creation"
fi

if command -v rpmbuild >/dev/null 2>&1; then
    create_rpm
fi

# Generate checksums
echo "🔐 Generating checksums..."
cd "$BUILD_DIR"
if ls *.tar.gz *.zip >/dev/null 2>&1; then
    sha256sum *.tar.gz *.zip > checksums.txt
    if ls *.deb >/dev/null 2>&1; then
        sha256sum *.deb >> checksums.txt
    fi
    if ls *.rpm >/dev/null 2>&1; then
        sha256sum *.rpm >> checksums.txt
    fi
else
    echo "No files to checksum"
fi
cd ..

echo "✅ Build complete!"
echo ""
echo "📁 Files created in $BUILD_DIR/:"
ls -la "$BUILD_DIR/"
echo ""
echo "🚀 To create a GitHub release:"
echo "   gh release create v$VERSION $BUILD_DIR/* --title \"Release v$VERSION\""
echo ""
echo "📦 To install locally:"
if [ -f "$BUILD_DIR/${APP_NAME}_${VERSION}_amd64.deb" ]; then
    echo "   sudo dpkg -i $BUILD_DIR/${APP_NAME}_${VERSION}_amd64.deb  # Ubuntu/Debian"
fi
if ls "$BUILD_DIR"/${APP_NAME}-${VERSION}-1.*.rpm >/dev/null 2>&1; then
    echo "   sudo rpm -i $BUILD_DIR/${APP_NAME}-${VERSION}-1.*.rpm     # CentOS/RHEL/Fedora"
fi
