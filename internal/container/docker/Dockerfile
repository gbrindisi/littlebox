# littlebox base image
# This image provides a secure sandbox environment for running AI coding agents.
# It includes network firewall capabilities and runs as a non-root user.

# Stage 1: Build libsandbox.so
FROM debian:bookworm-slim AS builder
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*
COPY libsandbox.c /tmp/libsandbox.c
RUN gcc -shared -fPIC -o /usr/local/lib/libsandbox.so /tmp/libsandbox.c -ldl

# Stage 2: Runtime image
FROM debian:bookworm-slim

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install required packages (runtime only, no build tools):
# - iptables: for network firewall rules
# - ipset: for efficient IP set management
# - iproute2: for ip command (network utilities)
# - dnsutils: for dig (domain resolution)
# - aggregate: for merging overlapping CIDR ranges
# - jq: for JSON parsing (GitHub IP ranges)
# - util-linux: for setpriv (privilege dropping)
# - ca-certificates: for HTTPS connections
# - git: commonly needed by coding agents
# - curl: for downloading tools and API calls
RUN apt-get update && apt-get install -y --no-install-recommends \
    iptables \
    ipset \
    iproute2 \
    dnsutils \
    aggregate \
    jq \
    util-linux \
    ca-certificates \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy compiled library from builder stage
COPY --from=builder /usr/local/lib/libsandbox.so /usr/local/lib/libsandbox.so

# Create the agent user (non-root)
# UID 1000 is commonly used and works well with bind mounts from the host
ARG AGENT_UID=1000
ARG AGENT_GID=1000
RUN groupadd -g ${AGENT_GID} agent && \
    useradd -m -u ${AGENT_UID} -g agent -s /bin/bash agent

# Copy scripts
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
COPY init-firewall.sh /usr/local/bin/init-firewall.sh
RUN chmod +x /usr/local/bin/entrypoint.sh /usr/local/bin/init-firewall.sh

# Create workspace directory
RUN mkdir -p /workspace && chown agent:agent /workspace

# Set working directory
WORKDIR /workspace

# Start as root for firewall initialization
# The entrypoint will drop privileges to the agent user after setup
USER root

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
