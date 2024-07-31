#!/bin/bash

if [[ "${NUON_LOCAL_CLI}" ]]; then
  echo "skipping installing nuon, and using local instance."
  exit 0
fi

echo "y" | /bin/bash -c "$(curl -fsSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)"
