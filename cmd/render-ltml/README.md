# render-ltml

`render-ltml` converts one or more LTML documents to PDF. By default, each input writes beside itself using the same filename with a `.pdf` extension. It can also submit the LTML and explicit uploaded assets to a remote `serve-ltml` instance instead of rendering locally.

Build with `-tags arabic` to enable the pure-Go Arabic shaper used by Arabic and
other complex-script LTML samples:

```sh
go build -tags arabic ./cmd/render-ltml
```

## Usage

```sh
render-ltml [flags] <file>
render-ltml -b [flags] <file1> <file2> ...
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `-assets <dir>` | `-a` | Directory of static assets available during rendering |
| `-extra <file>` | `-e` | Additional asset file (may be repeated) |
| `-output <path>` | `-o` | Output file in single-file mode, or output directory in batch mode |
| `-submit <url>` |  | Submit a multipart render request to this URL instead of rendering locally |
| `-watch` | `-w` | Watch inputs and assets for changes and rerender continuously |
| `-batch` | `-b` | Render multiple input files |

### Output paths

- When `-o` is omitted, each input writes to the same directory with its extension replaced by `.pdf`.
- Default output naming requires the input filename to have an extension.
- In single-file mode, `-o` names the PDF file to write.
- In batch mode, `-o` must name an existing output directory.
- If multiple batch inputs share the same basename and target the same output directory, later outputs overwrite earlier ones.

### Asset resolution

When `-assets` and/or `-extra` are given, a virtual filesystem is constructed and attached to the PDF writer before rendering. Asset-backed PDF operations resolve through this filesystem:

- Files supplied with `-extra` form the upper layer and shadow same-named files from `-assets`.
- Files in the `-assets` directory form the lower layer and are used when an asset is not supplied as an extra file.
- When neither flag is given, asset paths are resolved by the PDF writer directly relative to the working directory.

When an asset filesystem is attached, asset names must be clean relative `fs.FS` paths such as `logo.png` or `assets/logo.png`. Paths like `./logo.png`, `a/../logo.png`, or absolute paths are rejected.

If the same base name is given more than once via `-extra`, the last occurrence wins locally. Remote submission rejects duplicate `-extra` base names before sending the request.

### Watch mode

When `-watch` is set, `render-ltml` performs one render pass immediately and then keeps polling for changes:

- Input file changes rerender only the affected input.
- `-extra` file changes rerender all inputs that share those extra files.
- In local mode, changes anywhere under `-assets` rerender all inputs using that asset directory.
- In submit mode, `-watch` works for both single and batch operation.
- Render failures are reported, but watch mode keeps running for later changes.

### Remote submission mode

When `-submit` is set, `render-ltml` sends a `multipart/form-data` request to the given URL instead of rendering the PDF locally:

- Each LTML input file is sent as an `ltml` part in its own request.
- Each `-extra` file is uploaded as a `file` part.
- Each uploaded file uses its base name as the multipart `filename`, matching local `-extra` behavior.
- Batch mode submits one request per input file.
- `-assets` is not supported in remote mode; use explicit `-extra` uploads instead.

## Examples

Render one document beside the source LTML:

```sh
render-ltml report.ltml
```

Render to a specific file with a directory of shared assets:

```sh
render-ltml -a ./assets -o report.pdf report.ltml
```

Watch a single document and rerender on changes:

```sh
render-ltml -w report.ltml
```

Render a batch into one output directory:

```sh
render-ltml -b -o ./out reports/one.ltml reports/two.ltml
```

Submit a batch to a running `serve-ltml` instance:

```sh
render-ltml -b -submit http://localhost:8080/render -o ./out reports/one.ltml reports/two.ltml
```
