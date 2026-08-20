#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# go install drops the binary here, which isn't always already on PATH.
export PATH="$(go env GOPATH)/bin:$PATH"

VERSION=v2.0.1

# Both binaries ship: the CLI, and the runner embedded in the system image.
PACKAGES=(. ./gen)

# Packages go-licenses can't resolve a license for at all. Without these it
# fails whatever the policy says; the Apache-2.0 ones are attributed by hand
# further down.
IGNORE=(
  --ignore=github.com/alex-brook/composefork
  --ignore=gotest.tools
  --ignore=github.com/in-toto
)

if ! command -v go-licenses >/dev/null 2>&1; then
  go install "github.com/google/go-licenses/v2@$VERSION"
fi

rm -rf third_party THIRD-PARTY.csv

# The CSV is a readable index; third_party is what actually discharges the
# attribution, source-availability and NOTICE obligations.
go-licenses report "${PACKAGES[@]}" \
  --ignore=github.com/alex-brook/composefork > THIRD-PARTY.csv
go-licenses save "${PACKAGES[@]}" --save_path=third_party "${IGNORE[@]}"

for m in github.com/in-toto/in-toto-golang github.com/in-toto/attestation gotest.tools/v3; do
  src=$(find "$(go env GOMODCACHE)/$m"@* -maxdepth 0 -type d | head -1)
  mkdir -p "third_party/$m"
  cp "$src/LICENSE" "third_party/$m/LICENSE"
done
