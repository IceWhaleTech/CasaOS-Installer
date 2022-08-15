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

__warning() {
    echo -e "🟨 ${1}"
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

    __info "Downloading packages..."
    for PACKAGE in "${PACKAGES[@]}"; do
        curl -sLO "${PACKAGE}" || __error "Failed to download package"
    done

    __info "Extracting packages..."
    for PACKAGE_FILE in linux-*-casaos-*.tar.gz; do

        tar zxvf "${PACKAGE_FILE}" || __error "Failed to extract package"
    done

    BUILD_DIR=$(realpath -e "${TMP_DIR}"/build || __error "Failed to find build directory")

    popd
fi

SERVICES_TO_STOP=(
    "casaos.service"
    "casaos-gateway.service"
    "casaos-user-service.service"
)

__info "Stopping CasaOS services..."
for SERVICE in "${SERVICES_TO_STOP[@]}"; do
    systemctl stop "${SERVICE}" || __warning "Service ${SERVICE} does not exist."
done

MIGRATION_SCRIPT_DIR=$(realpath -e "${BUILD_DIR}"/scripts/migration/script.d || __error "Failed to find migration script directory")

__info "Running migration script before installation..."
for MIGRATION_SCRIPT in "${MIGRATION_SCRIPT_DIR}"/*.sh; do
    bash "${MIGRATION_SCRIPT}" || __error "Failed to run migration script"
done

__info "Installing CasaOS..."
SYSROOT_DIR=$(realpath -e "${BUILD_DIR}"/sysroot || __error "Failed to find sysroot directory")

# Generate manifest for uninstallation
MANIFEST_FILE=${BUILD_DIR}/sysroot/var/lib/casaos/manifest
touch "${MANIFEST_FILE}" || __error "Failed to create manifest file"
find "${SYSROOT_DIR}" -type f | cut -c ${#SYSROOT_DIR}- | cut -c 2- | tee "${MANIFEST_FILE}" || __error "Failed to create manifest file"

cp -rv "${SYSROOT_DIR}"/* / || __error "Failed to install CasaOS"

SETUP_SCRIPT_DIR=$(realpath -e "${BUILD_DIR}"/scripts/setup/script.d || __error "Failed to find setup script directory")

__info "Setting up CasaOS..."
for SETUP_SCRIPT in "${SETUP_SCRIPT_DIR}"/*.sh; do
    bash "${SETUP_SCRIPT}" || __error "Failed to run setup script"
done
