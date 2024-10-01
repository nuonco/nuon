#!/bin/bash

# Check if NUON_LOCAL_CLI is set
if [[ "${NUON_LOCAL_CLI}" ]]; then
  echo "Skipping installing nuon, and using local instance."
  exit 0
fi

echo "Installing nuon..."
echo "y" | /bin/bash -c "$(curl -fsSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)"

