#!/bin/sh
set -e

RELEASES_URL="https://github.com/Hootrix/go-mouse-keeper/releases"
BINARY_NAME="mouse-keeper"

# 检测系统和架构
detect_platform() {
    PLATFORM="$(uname -s)"
    case "${PLATFORM}" in
        Linux*)     PLATFORM="Linux";;
        Darwin*)    PLATFORM="Darwin";;
        *)          PLATFORM="unsupported"
    esac
    echo ${PLATFORM}
}

detect_arch() {
    ARCH="$(uname -m)"
    case "${ARCH}" in
        x86_64*)    ARCH="x86_64";;
        aarch64*)   ARCH="arm64";;
        *)          ARCH="unsupported"
    esac
    echo ${ARCH}
}

# 获取最新版本
get_latest_release() {
    curl --silent "https://api.github.com/repos/Hootrix/go-mouse-keeper/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                                                  # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                                         # Pluck JSON value
}

main() {
    PLATFORM=$(detect_platform)
    ARCH=$(detect_arch)
    
    if [ "$PLATFORM" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
        echo "Unsupported platform or architecture"
        exit 1
    fi
    
    VERSION=$(get_latest_release)
    ARCHIVE_NAME="${BINARY_NAME}_${PLATFORM}_${ARCH}.tar.gz"
    DOWNLOAD_URL="${RELEASES_URL}/download/${VERSION}/${ARCHIVE_NAME}"
    
    echo "Downloading ${BINARY_NAME} ${VERSION} for ${PLATFORM}_${ARCH}..."
    curl -sL "${DOWNLOAD_URL}" -o "${ARCHIVE_NAME}"
    
    echo "Installing ${BINARY_NAME}..."
    tar xzf "${ARCHIVE_NAME}"
    
    # 移动二进制文件到 /usr/local/bin 或 $HOME/.local/bin
    if [ -w "/usr/local/bin" ]; then
        mv "${BINARY_NAME}" "/usr/local/bin/"
    else
        mkdir -p "$HOME/.local/bin"
        mv "${BINARY_NAME}" "$HOME/.local/bin/"
        echo "Binary installed to $HOME/.local/bin"
        echo "Make sure to add $HOME/.local/bin to your PATH"
    fi
    
    # 清理下载的文件
    rm "${ARCHIVE_NAME}"
    
    echo "${BINARY_NAME} has been installed successfully!"
}

main
