# Releases

Leadtype is published as a Go module. A release consists of a semantic version
in `VERSION` and a matching Git tag. Pre-built command binaries are not
published.

## Creating a release

Choose the next version. The bump target updates `VERSION` and commits it with
a message such as `Bump version to 0.9.1.`:

```sh
make bump-patch # or bump-minor / bump-major
```

Verify the release without changing Git or publishing anything:

```sh
make release-check
```

Tag and publish the release:

```sh
make release
```

The `release` target requires all tracked changes to be committed and requires
an attached branch. Untracked files are allowed and are not included in the
release tag. The target validates `VERSION`, ensures the corresponding tag does
not exist locally or on `origin`, builds and tests the module, and verifies the
versions reported by `render-ltml` and `serve-ltml`. It then creates an
annotated `vVERSION` tag and atomically pushes the current branch and tag to
`origin`. If the push fails, the new local tag is removed.

## Depending on a release

A third-party program can select a Leadtype release with:

```sh
go get github.com/rowland/leadtype@v0.9.0
```

The selected version is recorded in the program's `go.mod`:

```go
require github.com/rowland/leadtype v0.9.0
```

Go also records module dependency versions in compiled programs. Inspect them
with:

```sh
go version -m path/to/program
```

All Leadtype packages currently share the version of the root Go module. A
future version 2 module must use the module path
`github.com/rowland/leadtype/v2`.
