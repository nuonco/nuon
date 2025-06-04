#!/bin/fish

### This script sets up a development environment on a fresh Arch Linux installation.
### It installs the necessary packages and tools for development, including AWS CLI, kubectl, and more.
### It also sets up the Nuon CLI and Twingate for VPN access.
### Make sure to run this script with sudo privileges.
### This script can be run on a fresh arch + fish system or a system with fish already installed.
### Usage: ./setup.sh


function setup_docker_insecure_registry
    # Function used to create a new docker daemon file with insecure registry for local devleopment
    # Usage: setup_docker_insecure_registry

    set DOCKER_CONFIG_FILE ~/.docker/daemon.json
    set INSECURE_REGISTRY "host.docker.internal:5001"

    # Check if the Docker config file exists
    if not test -f $DOCKER_CONFIG_FILE
        echo '{
    "insecure-registries": ["host.docker.internal:5001"]
}' >> $DOCKER_CONFIG_FILE
        echo "Docker config file created and insecure registry added."
    end
end


# install deps
set packages go aws-iam-authenticator aws-cli python39 python-pipx buf kubectl temporal-cli terraform-bin go-aws-sso nodejs k9s ngrok aws-nuke postgresql-libs twingate jq docker


for package in $packages
    if not yay -Qi $package > /dev/null
        echo "Installing $package..."
        yay -Sy --noconfirm $package
    else
        echo "$package is already installed."
    end
end

if not command -v aws-sso-util > /dev/null
    echo "aws-sso-util not found, installing..."
    pipx install aws-sso-util
else
    echo "aws-sso-util is already installed."
end

sudo usermod -aG docker $USER

echo "Do you want to setup twingate? (y/n): "
read user_input
if test "$user_input" = "y"
    sudo twingate setup
    sudo twingate start
else
    echo "Skipping Twingate setup."
end

# install nuon cli
if not command -v nuon > /dev/null
    echo "nuon not found, installing..."
    echo "y" | /bin/bash -c "(curl -fsSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/cli/install.sh)"
else
    echo "nuon is already installed."
end

if not command -v nuonctl > /dev/null
    echo "nuonctl not found, installing..."
    echo "y" | /bin/bash -c "(curl -fsSL https://nuon-artifacts.s3.us-west-2.amazonaws.com/nuonctl/install.sh)"
else
    echo "nuonctl is already installed."
end

# setup bin file for local binaries
mkdir -p ~/bin
mkdir -p ~/.local/bin
fish_add_path ~/bin ~/.local/bin



echo "Do you want to login into aws? (y/n): "
read user_input
if test "$user_input" = "y"
    nuonctl scripts exec init-aws
    aws-sso-util login
    set -x AWS_PROFILE stage.NuonAdmin
    set -x AWS_REGION us-west-2
else
    echo "skipping aws login"
end


echo "Do you want to monorepo dependencies? (y/n): "
read user_input
if test "$user_input" = "y"
    # install npm deps
    cd services/dashboard-ui
    npm install
    cd -

    # install deps for monorepo
    go mod download
else
    echo "Skipping dependency installation."
end

# code generation
echo "Do you want to reset generated code? (y/n): "
read user_input
if test "$user_input" = "y"
    nuonctl scripts exec reset-generated-code
else
    echo "Skipping code generation reset."
end

echo "Do you want to setup kubernetes ? (y/n): "
read user_input
if test "$user_input" = "y"
    echo "Setting up kubernetes..."
    nuonctl scripts exec init-kubernetes
    kx
    # test command, if this command return error then kubeconfig is not set up correctly
    kubectl get pods -n ctl-api
else
    echo "Skipping kubernetes setup."
end

setup_docker_insecure_registry

echo "Do you want to set ngrok token ? (y/n): "
read user_input
if test "$user_input" = "y"
    echo "Enter ngrok token: "
    read ngrok_token
    echo "Setting up ngrok token..."
    set -Ux NGROK_AUTHTOKEN $ngrok_token
else
    echo "Skipping ngrok token setup."
end

nuonctl scripts exec install-cli
