---
name: bump-go-deps
description: Bump this repo's Go dependencies, or diagnose a dependabot PR that fails to build. Use when a gomod dependabot PR is red, when a build reports "undefined:" inside a github.com/docker/cli or github.com/docker/docker package, or when asked to update or bump Go dependencies.
---

# Bumping Go dependencies

## Why this repo breaks in a way `go mod tidy` cannot fix

`github.com/docker/cli` and `github.com/docker/docker` are `+incompatible`:
neither ships a root `go.mod`, only a `vendor.mod` that Go ignores. That is
deliberate upstream — its header says a `go.mod` would opt them into SemVer,
and they ship CalVer. No `go.mod` is what makes them `+incompatible`, and
`+incompatible` is what makes Go treat them as having no requirements.

    go mod graph | grep -c '^github.com/docker/cli@'   # => 0

They contribute zero edges to the module graph, so MVS never learns what they
require. Their transitive deps float on whatever *other* modules pin — here
mostly `github.com/docker/compose/v5`.

A cli bump can therefore start calling a symbol our resolved version of some
dep does not have, with nothing in the graph forcing the upgrade. `go mod tidy`
does not catch it: tidy resolves and records versions, it does not typecheck.
The version it picks provides the *package*, just not the *symbol*.

Dependabot is blind to this for the same reason — it reads the same graph.

In practice only cli matters. We build a single package out of docker/docker —
`pkg/namesgenerator`, called once in `internal/cache.go` — and it imports
nothing outside stdlib, so it has no dependency versions that can skew. The
parity work below is deliberately cli-only. Generalise it if we ever import a
docker/docker package that has real dependencies of its own.

## Reproduce in ~1s

The embed tarballs are generated and gitignored, so a clean tree fails on
`//go:embed` before it typechecks — which disguises the real error. Stub them:

    touch internal/system_amd64.tar internal/system_arm64.tar
    go vet ./...

`go vet` rather than `go build`: it typechecks `_test.go` too, and `test/` is
almost entirely test files.

## Fixing an `undefined:` break

1. Read the symbol and the owning package out of the error.
2. `go get <dep>@<version that cli pins>` — see below.
3. `go mod tidy && go vet ./...`

## Aim for parity with docker/cli's vendored set

`vendor.mod` is the dependency combination upstream builds, tests and ships.
This project's premise is behavioural equivalence with docker compose, and the
compiler only proves the API lines up — not that behaviour does. So parity is
the target, and drift is worth closing even when everything compiles.

Parity is a **floor, not an exact pin**. MVS takes the maximum across the
graph, so where another module needs a dep newer than cli vendored, we sit
above it and that is correct:

    below cli's pin      -> close it
    at cli's pin         -> the target
    above, forced by MVS -> fine, leave it

Five sit above today — buildkit, containerd and compose force the prometheus
set and `go-winio`. Never `replace` them back down; those modules need what
they ask for.

Compare only modules the build actually needs. `go list -m all` includes
modules nothing imports, and `go mod why -m <mod>` reports "main module does
not need module" for those. `docker/go-events` and `go-jose/v4` are phantoms
today — `go mod tidy` will correctly drop any `go get` aimed at one.

## Checking parity

Reports every module we resolve *below* what cli vendored. Silence means parity.

    cli=$(go list -m -f '{{.Version}}' github.com/docker/cli)
    vendormod="$(go env GOMODCACHE)/github.com/docker/cli@$cli/vendor.mod"

    # Higher of two versions.
    higher() { printf '%s\n%s\n' "$1" "$2" | sort -V | tail -1; }

    # vendor.mod is go.mod syntax: each pin is an indented "module version"
    # pair, so drop comments and keep two-field lines with a v-prefixed second.
    sed 's|//.*||' "$vendormod" | awk 'NF==2 && $2 ~ /^v/ {print $1, $2}' |
    while read -r module pinned; do
      ours=$(go list -m -f '{{.Version}}' "$module" 2>/dev/null) || continue
      [ "$ours" = "$pinned" ] && continue                                       # at parity
      [ "$(higher "$ours" "$pinned")" = "$pinned" ] || continue                 # above, fine
      go mod why -m "$module" 2>/dev/null | grep -q 'does not need' && continue # phantom
      echo "below cli pin: $module ours=$ours cli=$pinned"
    done

The three `continue`s are the three reasons a difference is not a gap, in the
order given by the parity rule above: already matching, above because MVS
forced us there, or a module nothing in our build imports.

`sort -V` mis-ranks prereleases (`v1.1.0-rc.1` sorts after `v1.1.0`), and this
graph does contain them — `containerd/platforms v1.0.0-rc.5`. It errs toward
reporting a gap that is not one, so check anything surprising by hand.

## Verify

    go vet ./...
    go test ./internal/...   # ~2s
    go test ./test/          # ~8 min, needs docker
