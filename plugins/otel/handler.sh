#!/bin/sh
# Platform dispatch wrapper for claude-code-otel-plugin
# Detects OS and architecture, then execs the correct binary.

set -e

PLUGIN_ROOT="$(cd "$(dirname "$0")" && pwd)"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) exit 0 ;;
esac

BINARY="${PLUGIN_ROOT}/bin/handler-${OS}-${ARCH}"

if [ ! -x "$BINARY" ]; then
    exit 0
fi

exec "$BINARY"
