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

# Install CNI plugins for container networking (required for docker run)
ARG CNI_PLUGINS_VERSION=1.5.1
RUN mkdir -p /opt/cni/bin && \
    curl -fsSL https://github.com/containernetworking/plugins/releases/download/v${CNI_PLUGINS_VERSION}/cni-plugins-linux-amd64-v${CNI_PLUGINS_VERSION}.tgz -o /tmp/cni.tgz && \
    tar -xzf /tmp/cni.tgz -C /opt/cni/bin && \
    rm /tmp/cni.tgz

# Copy docker wrapper script
COPY docker-wrapper.sh /usr/local/bin/docker-wrapper.sh
RUN chmod +x /usr/local/bin/docker-wrapper.sh

# Remove existing docker and create symlink to wrapper
RUN if [ -f /usr/local/bin/docker ] || [ -L /usr/local/bin/docker ]; then \
        rm -f /usr/local/bin/docker; \
    fi && \
    ln -s /usr/local/bin/docker-wrapper.sh /usr/local/bin/docker && \
    chmod +x /usr/local/bin/nerdctl /usr/local/bin/docker

# Make executable
RUN chmod +x /usr/local/bin/nerdctl /usr/local/bin/docker

# Switch back to runner user (ARC expects this)
USER runner

# Final verification as runner user
RUN docker --version

# This image supports ALL docker commands:
# - docker build
# - docker push/pull
# - docker tag/images/rmi
# - docker run (with CNI networking)
# - docker exec/logs/ps/inspect
# - docker stop/rm
# - Multi-container scenarios