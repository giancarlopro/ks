FROM golang:1.23-bookworm

# The user matches the host uid and gid. That keeps the owner of every file in
# the bind-mounted workspace, and it removes the recursive chown of the home
# directory. That chown would walk the host's real ~/.claude, which is mounted.
ARG USERNAME=gian
ARG USER_UID=1000
ARG USER_GID=1000

# iptables, ipset, iproute2, and dnsutils are the firewall tooling.
# init-firewall.sh reads the container subnet with ip, and resolves each host
# with dig, so iproute2 and dnsutils are not optional.
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      curl \
      dnsutils \
      git \
      iproute2 \
      ipset \
      iptables \
      jq \
      make \
      sudo \
      zsh \
    && rm -rf /var/lib/apt/lists/*

# Node.js serves the Claude Code CLI through npm.
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# postCreateCommand and postStartCommand run as this user, and both need root.
RUN groupadd --gid "$USER_GID" "$USERNAME" \
    && useradd --uid "$USER_UID" --gid "$USER_GID" --shell /bin/bash --create-home "$USERNAME" \
    && echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/"$USERNAME" \
    && chmod 0440 /etc/sudoers.d/"$USERNAME"

# The Go build and module caches live in the user's home, so the user owns them.
RUN mkdir -p /home/"$USERNAME"/go \
    && chown -R "$USER_UID":"$USER_GID" /home/"$USERNAME"/go
ENV GOPATH=/home/gian/go
ENV PATH=/home/gian/go/bin:$PATH

USER $USERNAME
