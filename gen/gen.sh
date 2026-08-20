#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

BUILDER=multi
OUTDIR=./internal
PLATFORMS=(linux/amd64 linux/arm64)

# --- idempotent builder: create only if it doesn't already exist ---
if ! docker buildx inspect "$BUILDER" >/dev/null 2>&1; then
  docker buildx create --name "$BUILDER" --driver docker-container --bootstrap
fi
docker buildx use "$BUILDER"

# --- one single-arch image tarball per platform ---
for platform in "${PLATFORMS[@]}"; do
  arch="${platform##*/}"   # linux/amd64 -> amd64
  docker buildx build \
    --platform "$platform" \
    -f gen/Dockerfile \
    -o "type=docker,dest=${OUTDIR}/system_${arch}.tar" \
    -t composefork/system \
    .
done
