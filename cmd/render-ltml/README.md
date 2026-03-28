# render-ltml

`render-ltml` converts an LTML document to PDF and writes the result to a file or stdout. It can also submit the LTML and explicit uploaded assets to a remote `serve-ltml` instance instead of rendering locally.

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
| `-submit <url>` |  | Submit a multipart render request to this URL instead of rendering locally |

### Asset resolution

When `-assets` and/or `-extra` are given, a virtual filesystem is constructed and attached to the PDF writer before rendering. Asset-backed PDF operations resolve through this filesystem:

- Files supplied with `-extra` form the **upper layer** and shadow same-named files from `-assets`.
- Files in the `-assets` directory form the **lower layer** and are used when an asset is not supplied as an extra file.
- When neither flag is given, asset paths are resolved by the PDF writer directly (relative to the working directory).

When an asset filesystem is attached, asset names must be clean relative `fs.FS` paths such as `logo.png` or `assets/logo.png`. Paths like `./logo.png`, `a/../logo.png`, or absolute paths are rejected.

If the same base name is given more than once via `-extra`, the last occurrence wins.

### Remote submission mode

When `-submit` is set, `render-ltml` sends a `multipart/form-data` request to the given URL instead of rendering the PDF locally:

- The LTML input file is sent as the first `ltml` part.
- Each `-extra` file is uploaded as a `file` part.
- Each uploaded file uses its base name as the multipart `filename`, matching local `-extra` behavior.
- Duplicate `-extra` base names are rejected before the request is sent.
- `-assets` is not supported in remote mode; use explicit `-extra` uploads instead.

The command still writes the returned PDF to stdout or `-output`.

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

Submit a document to a running `serve-ltml` instance and write the returned PDF to a file:

```sh
render-ltml -submit http://localhost:8080/render -o report.pdf report.ltml
```

Submit a document plus one explicit uploaded asset:

```sh
render-ltml -submit http://localhost:8080/render \
  -e ./branded/logo.png \
  -o report.pdf \
  report.ltml
```
