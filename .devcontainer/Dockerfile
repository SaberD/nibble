FROM golang:1.26

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

# Chromium in its own layer for better cache reuse.
RUN apt-get update && apt-get install -y \
    --no-install-recommends chromium

# Base system utilities and project dependencies.
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    fonts-dejavu-core \
    fonts-noto-color-emoji \
    git \
    libpcap-dev \
    ffmpeg \
    npm \
    python3-pip \
    sudo

# VHS runtime tools.
RUN set -eux; \
    arch="$(dpkg --print-architecture)"; \
    if [ "$arch" = "amd64" ]; then ttyd_arch="x86_64"; \
    elif [ "$arch" = "arm64" ]; then ttyd_arch="aarch64"; \
    else echo "Unsupported architecture: $arch (expected amd64 or arm64)" >&2; exit 1; fi; \
    curl -fsSL "https://github.com/tsl0922/ttyd/releases/latest/download/ttyd.${ttyd_arch}" -o /usr/local/bin/ttyd; \
    chmod +x /usr/local/bin/ttyd; \
    go install github.com/charmbracelet/vhs@latest; \
    cp /go/bin/vhs /usr/local/bin/vhs

# dev user
RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} -m -s /bin/bash ${USERNAME} \
    && usermod -aG sudo ${USERNAME} \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} \
    && chmod 440 /etc/sudoers.d/${USERNAME}

# goreleaser
RUN curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar xz \
    && mv goreleaser /usr/local/bin/goreleaser

# Ensure Go workspace paths are writable by dev user.
RUN mkdir -p /go/pkg /go/bin \
    && chown -R ${USERNAME}:${USERNAME} /go

USER ${USERNAME}
WORKDIR /workspace
