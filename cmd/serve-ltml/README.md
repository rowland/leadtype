# serve-ltml

`serve-ltml` is an HTTP server that renders LTML documents to PDF on demand. Clients submit an LTML document and optional asset files in a single `multipart/form-data` request and receive a PDF response.

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
| `-max-upload-bytes <n>` | `MAX_UPLOAD_BYTES` | `33554432` (32 MiB) | Maximum request body size |
| `-read-timeout <duration>` | `READ_TIMEOUT` | none | HTTP server read timeout (e.g. `30s`) |
| `-write-timeout <duration>` | `WRITE_TIMEOUT` | none | HTTP server write timeout (e.g. `60s`) |

`ASSETS` must exist and be a directory; the server refuses to start otherwise.

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

## Examples

Start the server:

```sh
serve-ltml -assets /var/lib/ltml/assets
```

Or with environment variables:

```sh
ASSETS=/var/lib/ltml/assets READ_TIMEOUT=30s WRITE_TIMEOUT=60s serve-ltml
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

Place an asset at a nested path:

```sh
curl -s \
  -F 'ltml=@report.ltml;type=application/vnd.rowland.leadtype.ltml+xml' \
  -F 'file=@./img/logo.png;filename=assets/logo.png' \
  http://localhost:8080/render -o report.pdf
```
