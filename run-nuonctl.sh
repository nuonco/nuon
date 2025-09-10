#!/bin/bash

# Exit on error
set -e

DD_API_KEY=758a582665d6506f1420e5680925a091
DD_APP_KEY=284530e2c85aa36ed529b125c269d4b986883f08

# write metrics to datadog for which user is using the script, and which command is being used.
function write_metrics() {
  # Set default email to "na" in case of failure
  local email="na"
  
  # Attempt to get AWS caller identity, but don't fail if it doesn't work
  set +e
  local aws_information=$(aws sts get-caller-identity --no-cli-pager 2>/dev/null)
  local aws_exit_code=$?
  set -e

  # Try to extract email if aws command was successful
  if [ $aws_exit_code -eq 0 ]; then
    email=$(echo "$aws_information" | jq -r '.UserId | split(":")[1] // "na"')
  fi
  
  # Default values
  local namespace="default"
  local subcommand="unknown"
  local full_command="unknown"
  local flags_tags=""
  local args_tags=""

  # Parse arguments more intelligently
  if [ $# -gt 0 ]; then
    # First argument is the namespace/service
    namespace="${1}"
    shift

    # Second argument is the subcommand
    if [ $# -gt 0 ]; then
      subcommand="${1}"
      shift
    fi

    # Full command is namespace + subcommand
    full_command="${namespace} ${subcommand}"

    # Process remaining arguments as flags and args
    local flags=()
    local non_flags=()
    for arg in "$@"; do
      if [[ "$arg" == --* ]]; then
        flags+=("$arg")
      else
        non_flags+=("$arg")
      fi
    done

    # Create tags for flags
    if [ ${#flags[@]} -gt 0 ]; then
      flags_tags=$(printf "flags:%s," "${flags[@]}" | sed 's/,$//')
    fi

    # Create tags for non-flag args
    if [ ${#non_flags[@]} -gt 0 ]; then
      args_tags=$(printf "args:%s," "${non_flags[@]}" | sed 's/,$//')
    fi
  fi

  local timestamp=$(date +%s)

  # Prepare the metrics payload
  local metrics_payload=$(cat << EOF
{
  "series": [
    {
      "metric": "nuonctl.run",
      "type": 3,
      "points": [
        {
          "timestamp": $timestamp,
          "value": 1
        }
      ],
      "resources": [
        {"name":"nuonctl"},
        {"type":"service"}
      ],
      "tags": [
        "service:nuonctl",
        "email:$email",
        "namespace:$namespace", 
        "subcommand:$subcommand",
        "full_command:$full_command",
        "${flags_tags}${args_tags}flags_count:${#flags[@]}",
        "args_count:${#non_flags[@]}"
      ]
    }
  ]
}
EOF
)

  # Debug logging if NUON_DEBUG is set to true
  if [[ "${NUON_DEBUG:-}" == "true" ]]; then
    echo "DataDog Metrics Payload:" >&2
    echo "$metrics_payload" >&2
  fi

  # Send metrics to DataDog
  # Capture curl output
  local curl_response=$(curl -X POST "https://us5.datadoghq.com/api/v2/series" \
  -s \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -d "$metrics_payload")

  # Check for errors using jq
  local errors=$(echo "$curl_response" | jq -r '.errors // empty | select(length > 0)')
  
  # Print errors if they exist
  if [ -n "$errors" ]; then
    echo "Error sending metrics to DataDog: $errors" >&2
  fi
  echo  >&2 "successfully wrote usage metrics"
}

echo  >&2 "writing metrics to run-nuonctl"
write_metrics $@

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

  export RUN_NUONCTL_VERSION=local
  export RUN_NUONCTL_PATH=$EXEC_PATH
  exec_and_cleanup $EXEC_PATH $@
else
    # Use the install script for the regular flow
    # Check if install script exists
    if [ ! -f "$INSTALL_SCRIPT" ]; then
        echo "❌ Error: Installation script not found at $INSTALL_SCRIPT"
        exit 1
    fi

    if [ -f "$INSTALL_DIR/nuonctl" ]; then
      export RUN_NUONCTL_VERSION="check-version"
      export RUN_NUONCTL_PATH="check-version"
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

    export RUN_NUONCTL_VERSION=$LATEST_VERSION
    export RUN_NUONCTL_PATH=$EXEC_PATH
    exec_and_cleanup $EXEC_PATH $@
fi
