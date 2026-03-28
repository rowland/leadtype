# render-ltml

`render-ltml` converts an LTML document to PDF and writes the result to a file or stdout.

## Usage

```
render-ltml [flags] <file.ltml>
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `-assets <dir>` | `-a` | Directory of static assets available during rendering |
| `-extra <file>` | `-e` | Additional asset file (may be repeated) |
| `-output <file>` | `-o` | Write PDF to this file instead of stdout |

### Asset resolution

When `-assets` and/or `-extra` are given, a virtual filesystem is constructed and attached to the PDF writer before rendering. Asset-backed PDF operations resolve through this filesystem:

- Files supplied with `-extra` form the **upper layer** and shadow same-named files from `-assets`.
- Files in the `-assets` directory form the **lower layer** and are used when an asset is not supplied as an extra file.
- When neither flag is given, asset paths are resolved by the PDF writer directly (relative to the working directory).

When an asset filesystem is attached, asset names must be clean relative `fs.FS` paths such as `logo.png` or `assets/logo.png`. Paths like `./logo.png`, `a/../logo.png`, or absolute paths are rejected.

If the same base name is given more than once via `-extra`, the last occurrence wins.

## Examples

Render to stdout and pipe into a PDF viewer:

```sh
render-ltml report.ltml | zathura -
```

Render to a file with a directory of shared assets:

```sh
render-ltml -a ./assets -o report.pdf report.ltml
```

Override one asset without touching the shared directory:

```sh
render-ltml -a ./assets -e ./branded/logo.png -o report.pdf report.ltml
```
