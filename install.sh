#!/usr/bin/bash

set -e

BUILD_DIR=${1}

TMP_ROOT=/tmp/casaos-installer

__info() {
    echo -e "🟩 ${1}"
}

__info_done() {
    echo -e "✅ ${1}"
}

__error() {
    echo "🟥 ${1}"
    exit 1
}

OS=$(uname || "unknown")

if [ "${OS}" != "Linux" ]; then
    echo "This script is only for Linux"
    exit 1
fi

ARCH="unknown"

case $(uname -m) in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    armv7l)
        ARCH="arm-7"
        ;;
    *)
        echo "Unsupported architecture"
        exit 1
        ;;
esac

if [ -z "${BUILD_DIR}" ]; then

    PACKAGES=(
        "https://github.com/IceWhaleTech/CasaOS-Gateway/releases/download/v0.3.5-alpha7/linux-${ARCH}-casaos-gateway-v0.3.5-alpha7.tar.gz"
        "https://github.com/IceWhaleTech/CasaOS-UserService/releases/download/v0.3.5-alpha3/linux-${ARCH}-casaos-user-service-migration-tool-v0.3.5-alpha3.tar.gz"
    )

    mkdir -p ${TMP_ROOT} || __error "Failed to create temporary directory"
    TMP_DIR=$(mktemp -d -p ${TMP_ROOT} || __error "Failed to create temporary directory")

    pushd "${TMP_DIR}"

    for PACKAGE in "${PACKAGES[@]}"; do
        __info "Downloading ${PACKAGE}..."
        curl -sLO "${PACKAGE}" || __error "Failed to download package"
    done

    for PACKAGE_FILE in linux-*-casaos-*.tar.gz; do
        __info "Extracting ${PACKAGE_FILE}..."
        tar zxvf "${PACKAGE_FILE}" || __error "Failed to extract package"
    done

    BUILD_DIR=$(realpath -e "${TMP_DIR}"/build || __error "Failed to find build directory")

    popd
fi

MIGRATION_SCRIPT_DIR=$(realpath -e "${BUILD_DIR}"/scripts/migration/script.d || __error "Failed to find migration script directory")

for MIGRATION_SCRIPT in "${MIGRATION_SCRIPT_DIR}"/*.sh; do
    __info "Running ${MIGRATION_SCRIPT}..."
    bash "${MIGRATION_SCRIPT}" || __error "Failed to run migration script"
done

