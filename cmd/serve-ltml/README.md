# serve-ltml

`serve-ltml` is an HTTP server that renders LTML documents to PDF on demand. Clients submit an LTML document and optional asset files in a single `multipart/form-data` request and receive a PDF response.

Arabic and other complex-script text use the built-in pure-Go shaper, so normal server builds render shaped right-to-left output.

## Usage

```
serve-ltml [flags]
```

Every flag can also be set via the corresponding environment variable.

### Flags and environment variables

| Flag | Environment variable | Default | Description |
|------|----------------------|---------|-------------|
| `-listen <addr>` | `LISTEN` | `:8080` | Address to listen on |
| `-assets <dir>` / `-a <dir>` | `ASSETS` | *(required)* | Directory of static assets available to all requests |
| `-font-dir <dirs>` | `FONT_DIR` | `auto` | Ordered comma-delimited font directories; `auto` inserts system directories |
| `-output-path <dir>` | `OUTPUT_PATH` | none | Existing root directory for render output persisted on behalf of requests carrying `X-Output-File` |
| `-max-upload-bytes <n>` | `MAX_UPLOAD_BYTES` | `33554432` (32 MiB) | Maximum request body size |
| `-read-timeout <duration>` | `READ_TIMEOUT` | none | HTTP server read timeout (e.g. `30s`) |
| `-write-timeout <duration>` | `WRITE_TIMEOUT` | none | HTTP server write timeout (e.g. `60s`) |
| `-version` |  |  | Print the compiled Leadtype version and exit |

`ASSETS` must exist and be a directory; the server refuses to start otherwise.

Font directories are searched in the order configured. Use `./fonts` for a
deterministic custom-only inventory, `./fonts,auto` to prefer custom fonts and
then use system fonts, or `auto,./fonts` to prefer system fonts. Explicit
directories must exist; unavailable directories introduced by `auto` are
ignored.

## API

### `POST /render`

Render an LTML document to PDF.

**Request**

`Content-Type: multipart/form-data`

| Part | Field name | Required | Description |
|------|------------|----------|-------------|
| LTML document | `ltml` | Yes | Must be the **first** part. Preferred content type: `application/vnd.rowland.leadtype.ltml+xml`; `application/xml`, `text/xml`, and no content type are also accepted. |
| Asset file | `file` | No | May be repeated. The part's `filename` parameter is used as the virtual asset path (e.g. `logo.png` or `assets/logo.png`). |

**Response**

| Status | Meaning |
|--------|---------|
| `200 OK` | PDF rendered successfully. Body is the PDF; `Content-Type: application/pdf`; `Content-Disposition: inline; filename="output.pdf"`. |
| `400 Bad Request` | Malformed multipart body, missing or misplaced `ltml` part, empty LTML, or invalid upload filename. |
| `405 Method Not Allowed` | Request method is not `POST`. |
| `413 Request Entity Too Large` | Request body exceeds `-max-upload-bytes`. |
| `500 Internal Server Error` | Temp-file, parse, render, or stream failure. |

### Asset resolution

Uploaded files form a **per-request upper layer** that shadows same-named files in the configured assets directory for the duration of that request only. Parallel requests never share upload state. Uploaded filenames must be clean relative `fs.FS` paths such as `logo.png` or `assets/logo.png`; empty names, `.`, paths containing `.` / `..` segments, and absolute paths are rejected.

### Persisting render output

By default, a successful `/render` request streams the generated PDF in the
response body. Configuring `-output-path` enables an alternative, request-driven
file output mode; it does not change the behavior of requests by itself.

To request persisted output, the client sends an `X-Output-File` header:

```http
X-Output-File: reports/report.pdf
```

The header value is a relative output filename. Nested paths are allowed, but
absolute paths, `.` and `..` path segments, and unclean paths are rejected.
For each successful request, the server creates a new ULID-named directory
under `-output-path` and writes:

```text
<output-path>/
└── <ULID>/
    ├── reports/
    │   ├── report.pdf
    │   └── report.ltml
    └── <uploaded files at their multipart paths>
```

The `.ltml` file contains the submitted source. Every uploaded `file` part is
also retained at its multipart filename, relative to the same ULID directory.
An upload may not use the path reserved for either the generated PDF or its
LTML sidecar.

File output mode returns JSON instead of PDF bytes:

```json
{
  "path": "01JQXYZ123456789ABCDEFGHJK/reports/report.pdf",
  "size": 42817,
  "elapsed_ms": 36
}
```

`path` is relative to the configured output directory, `size` is the generated
PDF size in bytes, and `elapsed_ms` covers rendering through writing the PDF.
Errors are also returned as JSON with `error` and `elapsed_ms` fields.

The configured output directory must already exist when the server starts, and
the server process must be able to create directories and files beneath it.
The server does not apply a retention policy or delete successful persisted
outputs. If `X-Output-File` is sent without `-output-path`/`OUTPUT_PATH`
configured, the request fails with `500 Internal Server Error`.

## Examples

Start the server:

```sh
serve-ltml -assets /var/lib/ltml/assets

# Container-friendly: search only the bundled inventory.
serve-ltml -assets /var/lib/ltml/assets -font-dir /var/lib/ltml/fonts
```

Start the server with persisted output enabled:

```sh
serve-ltml \
  -assets /var/lib/ltml/assets \
  -output-path /var/lib/ltml/output
```

Or with environment variables:

```sh
ASSETS=/var/lib/ltml/assets FONT_DIR=/var/lib/ltml/fonts READ_TIMEOUT=30s WRITE_TIMEOUT=60s serve-ltml
```

Render a document with no uploaded assets:

```sh
curl -s \
  -F 'ltml=@report.ltml;type=application/vnd.rowland.leadtype.ltml+xml' \
  http://localhost:8080/render -o report.pdf
```

Render with an asset that overrides the server's configured asset copy:

```sh
curl -s \
  -F 'ltml=@report.ltml;type=application/vnd.rowland.leadtype.ltml+xml' \
  -F 'file=@./branded/logo.png;filename=logo.png' \
  http://localhost:8080/render -o report.pdf
```

Persist a render and receive its server-side path as JSON:

```sh
curl -sS \
  -H 'X-Output-File: reports/report.pdf' \
  -F 'ltml=@report.ltml;type=application/vnd.rowland.leadtype.ltml+xml' \
  -F 'file=@./img/logo.png;filename=assets/logo.png' \
  http://localhost:8080/render
```

Place an asset at a nested path:

```sh
curl -s \
  -F 'ltml=@report.ltml;type=application/vnd.rowland.leadtype.ltml+xml' \
  -F 'file=@./img/logo.png;filename=assets/logo.png' \
  http://localhost:8080/render -o report.pdf
```


## Embedding and custom widgets

`serve-ltml` now exposes an importable package at:

- `github.com/rowland/leadtype/cmd/serve-ltml/serveltml`

Third-party programs can host the server and register custom LTML widgets
before requests are handled:

```go
code := serveltml.Main(os.Stderr, func() error {
    // call ltml.RegisterTag / custom setup here
    return nil
})
os.Exit(code)
```
