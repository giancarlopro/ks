#!/usr/bin/env bash
# Deny egress by default. Allow the hosts in ALLOWED_DOMAINS.
#
# The list resolves once, when this script runs. A host that changes address
# needs another run. A host behind a CDN with names that are not fixed cannot be
# allowed at all: run sandbox-firewall-off, do the one task, and run this again.
set -euo pipefail

ALLOWED_DOMAINS=(
  # Claude Code itself. The last one carries error reports.
  api.anthropic.com
  claude.ai
  console.anthropic.com
  sentry.io
  # The npm registry serves the Claude Code CLI.
  registry.npmjs.org
  github.com
  api.github.com
  codeload.github.com
  objects.githubusercontent.com
  # Debian bookworm packages.
  deb.debian.org
  security.debian.org
  # Go modules. The proxy redirects module zips to the object store, so the
  # object store is needed too.
  proxy.golang.org
  sum.golang.org
  storage.googleapis.com
)

# Reset to a known-open state first. Flushing leaves the default policies alone,
# so a second run would otherwise inherit DROP and starve its own dig lookups.
iptables -P INPUT ACCEPT
iptables -P FORWARD ACCEPT
iptables -P OUTPUT ACCEPT
iptables -F
iptables -X

# The set survives a flush, so reuse it rather than depending on destroy.
ipset create allowed hash:net -exist
ipset flush allowed

# DNS, loopback, and established traffic come first.
iptables -A OUTPUT -o lo -j ACCEPT
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
iptables -A OUTPUT -p tcp --dport 53 -j ACCEPT
iptables -A OUTPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# The host reaches the container through the default gateway. Keep that subnet.
# Read the real CIDR off the link route -- the docker bridge is commonly a /16,
# so assuming a /24 around the gateway drops part of it.
HOST_NET=$(ip route | awk '/proto kernel/ && /src/ {print $1; exit}')
if [[ -z ${HOST_NET:-} ]]; then
  echo "init-firewall: could not read the container subnet; leaving the firewall down" >&2
  exit 1
fi
iptables -A OUTPUT -d "$HOST_NET" -j ACCEPT
iptables -A INPUT -s "$HOST_NET" -j ACCEPT

for domain in "${ALLOWED_DOMAINS[@]}"; do
  for ip in $(dig +short A "$domain"); do
    [[ $ip =~ ^[0-9.]+$ ]] && ipset add allowed "$ip" -exist
  done
done

iptables -A OUTPUT -m set --match-set allowed dst -j ACCEPT
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT DROP

echo "firewall up: ${#ALLOWED_DOMAINS[@]} domains allowed"
