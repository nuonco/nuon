FROM ghcr.io/actions/actions-runner:latest

USER root

# Install nerdctl (latest version)
ARG NERDCTL_VERSION=2.1.6
RUN curl -sSL https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-amd64.tar.gz | \
    tar -xz -C /usr/local/bin/ nerdctl

# Install BuildKit client (buildctl only, not buildkitd daemon)
# Match version with your BuildKit server (v0.24.0)
ARG BUILDKIT_VERSION=0.24.0
RUN curl -sSL https://github.com/moby/buildkit/releases/download/v${BUILDKIT_VERSION}/buildkit-v${BUILDKIT_VERSION}.linux-amd64.tar.gz | \
    tar -xz -C /tmp && \
    mv /tmp/bin/buildctl /usr/local/bin/ && \
    rm -rf /tmp/bin && \
    chmod +x /usr/local/bin/buildctl

# Install docker buildx CLI plugin
ARG BUILDX_VERSION=0.19.3
RUN mkdir -p /usr/local/lib/docker/cli-plugins && \
    curl -fsSL https://github.com/docker/buildx/releases/download/v${BUILDX_VERSION}/buildx-v${BUILDX_VERSION}.linux-amd64 \
    -o /usr/local/lib/docker/cli-plugins/docker-buildx && \
    chmod +x /usr/local/lib/docker/cli-plugins/docker-buildx

# Copy docker wrapper script
COPY docker-wrapper.sh /usr/local/bin/docker-wrapper.sh
RUN chmod +x /usr/local/bin/docker-wrapper.sh

# Remove existing docker and create symlink to wrapper
RUN if [ -f /usr/local/bin/docker ] || [ -L /usr/local/bin/docker ]; then \
        rm -f /usr/local/bin/docker; \
    fi && \
    ln -s /usr/local/bin/docker-wrapper.sh /usr/local/bin/docker && \
    chmod +x /usr/local/bin/nerdctl /usr/local/bin/docker


# Switch back to runner user (ARC expects this)
USER runner

# Final verification as runner user
RUN docker --version

# This image supports:
# - docker build
# - docker push
# - docker pull
# - docker tag
# - docker images
# - docker rmi
# Does NOT support docker run (no CNI plugins)