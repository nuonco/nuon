#!/bin/bash

# Exit on error
set -e

# Directory definitions
MONO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NUONCTL_DIR="${MONO_ROOT}/bins/nuonctl"
INSTALL_SCRIPT="${NUONCTL_DIR}/install.sh"
BASE_URL=https://nuon-artifacts.s3.us-west-2.amazonaws.com/nuonctl

# Install to user's bin directory
INSTALL_DIR="$HOME/bin"
if [ ! -d "$INSTALL_DIR" ]; then
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    elif [ -w "/bin" ]; then
        INSTALL_DIR="/bin"
    else
        echo "❌ Cannot find a writable bin directory. Try running with sudo."
        exit 1
    fi
fi

function install() {
    # Make sure install.sh is executable
    chmod +x "$INSTALL_SCRIPT"

    # Run the installation script
    echo "Running nuonctl installation script..."
    echo "y" | bash "$INSTALL_SCRIPT"

    # Check if installation was successful
    if [ $? -eq 0 ]; then
        echo "🎉 nuonctl installation completed successfully!"
    else
        echo "❌ nuonctl installation failed. Please check the error messages above."
        exit 1
    fi
}

# Check if we should build locally
if [ -n "${NUON_LOCAL:-}" ]; then
    echo "🛠️  NUON_LOCAL is set, building nuonctl locally..."

    # Check for Go installation
    if ! command -v go &> /dev/null; then
        echo "❌ Error: Go is required to build nuonctl locally but it's not installed."
        exit 1
    fi

    # Assuming the nuonctl source is in the bins/nuonctl directory
    NUONCTL_SRC_DIR="${MONO_ROOT}/bins/nuonctl"

    echo "Building from source in ${NUONCTL_SRC_DIR}..."
    cd "${NUONCTL_SRC_DIR}"

    # Build the binary
    go build -o nuonctl-local .


    echo "Installing nuonctl to ${INSTALL_DIR}..."
    mv nuonctl-local "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/nuonctl-local"

    echo "🎉 nuonctl local build and installation completed successfully!"
else
    # Use the install script for the regular flow
    # Check if install script exists
    if [ ! -f "$INSTALL_SCRIPT" ]; then
        echo "❌ Error: Installation script not found at $INSTALL_SCRIPT"
        exit 1
    fi

    if [ -f "$INSTALL_DIR/nuonctl" ]; then
      CURRENT_VERSION=`exec $INSTALL_DIR/nuonctl version`
      echo " > CURRENT_VERSION: $CURRENT_VERSION"

      echo "calculating latest version..."
      LATEST_VERSION=$(curl -s $BASE_URL/latest.txt)
      echo " > LATEST_VERSION: $LATEST_VERSION"

      if [[ "$CURRENT_VERSION" == "$LATEST_VERSION" ]]; then
        echo "currently installed version is current - doing nothing"
      else
        install
      fi
    else
      install
    fi
fi


if [ -n "${NUON_LOCAL:-}" ]; then
  exec ${INSTALL_DIR}/nuonctl-local $@
else
  exec ${INSTALL_DIR}/nuonctl $@
fi
