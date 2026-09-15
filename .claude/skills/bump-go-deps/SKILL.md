---
name: bump-go-deps
description: Bump this repo's Go dependencies, or diagnose a dependabot PR that fails to build. Use when a gomod dependabot PR is red, when a build reports "undefined:" inside a github.com/docker/cli or github.com/docker/docker package, or when asked to update or bump Go dependencies.
---

# Bumping Go dependencies

## Why this repo breaks in a way `go mod tidy` cannot fix

`github.com/docker/cli` and `github.com/docker/docker` are `+incompatible`:
neither ships a root `go.mod`, only a `vendor.mod` that Go ignores.

    go mod graph | grep -c '^github.com/docker/cli@'   # => 0

They contribute zero edges to the module graph, so MVS never learns what they
require. Their transitive deps float on whatever *other* modules pin — here
mostly `github.com/docker/compose/v5`.

A cli bump can therefore start calling a symbol our resolved version of some
dep does not have, with nothing in the graph forcing the upgrade. `go mod tidy`
does not catch it: tidy resolves and records versions, it does not typecheck.
The version it picks provides the *package*, just not the *symbol*.

Dependabot is blind to this for the same reason — it reads the same graph.

## Reproduce in ~1s

The embed tarballs are generated and gitignored, so a clean tree fails on
`//go:embed` before it typechecks — which disguises the real error. Stub them:

    touch internal/system_amd64.tar internal/system_arm64.tar
    go build ./...

## Fixing an `undefined:` break

1. Read the symbol and the owning package out of the error.
2. Find the version that introduced it. The pinned cli's own pin is the best
   first guess:

       grep <dep> "$(go env GOMODCACHE)"/github.com/docker/cli@<ver>+incompatible/vendor.mod

3. `go get <dep>@<version>`
4. `go mod tidy && go build ./... && go vet ./...`

## Do not mass-align against vendor.mod

Tempting and wrong. `vendor.mod` is cli's *vendored snapshot*, not its minimum
requirements. Sitting below those pins is the normal steady state: `main` has
sat below ~15 of them while fully green, and breaks only when cli actually
calls a newer symbol. A check that flags every gap is red on a healthy tree,
which is worse than no check.

The compiler is the oracle. Bump what fails to build. Bumping neighbours to
match cli is optional tidiness, not a fix — say which it is.

## Verify

    go build ./... && go vet ./...
    go test ./internal/...   # ~2s
    go test ./test/          # ~8 min, needs docker
