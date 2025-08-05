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

function build_locally() {
    echo >&2 "🛠️  NUONCTL_LOCAL is set, building nuonctl locally..."

    # Check for Go installation
    if ! command -v go &> /dev/null; then
        echo "❌ Error: Go is required to build nuonctl locally but it's not installed."
        exit 1
    fi

    # Assuming the nuonctl source is in the bins/nuonctl directory
    NUONCTL_SRC_DIR="${MONO_ROOT}/bins/nuonctl"

    echo >&2 "Building from source in ${NUONCTL_SRC_DIR}..."
    cd "${NUONCTL_SRC_DIR}"

    TMP_PATH=/tmp/nctl-$(date +%s)
    go build -o $TMP_PATH .

    echo >&2 "🎉 nuonctl local build and installation completed successfully!"

    echo $TMP_PATH
}

function exec_and_cleanup() {
  local exec_path="$1"
  shift

  echo >&2 "executing $exec_path..."

  exec "$exec_path" "$@"
  rm -f "$exec_path"
}

# Check if we should build locally
if [ "${NUONCTL_LOCAL:-}" = "true" ]; then
  EXEC_PATH=$(build_locally)
  exec_and_cleanup $EXEC_PATH $@
else
    # Use the install script for the regular flow
    # Check if install script exists
    if [ ! -f "$INSTALL_SCRIPT" ]; then
        echo "❌ Error: Installation script not found at $INSTALL_SCRIPT"
        exit 1
    fi

    if [ -f "$INSTALL_DIR/nuonctl" ]; then
      CURRENT_VERSION=`exec $INSTALL_DIR/nuonctl version`
      echo "calculating latest version..."
      LATEST_VERSION=$(curl -s $BASE_URL/latest.txt)

      if [[ "$CURRENT_VERSION" == "$LATEST_VERSION" ]]; then
        echo  >&2 "currently installed version is current ($LATEST_VERSION) - doing nothing"
      else
        echo  >&2 "currently installed version is out of date - installing"
        install
      fi
    else
      echo  >&2 "nuonctl is not currently installed - installing"
      install
    fi

    # NOTE(jm): we do this, so that long running nuonctl processes do not have to be terminated when an update comes in.
    EXEC_PATH=/tmp/nuonctl-$(date +%s)
    cp $INSTALL_DIR/nuonctl $EXEC_PATH 

    exec_and_cleanup $EXEC_PATH $@
fi
