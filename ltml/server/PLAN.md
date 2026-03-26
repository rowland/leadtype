# LTML Multipart Render Server Plan

This document is the implementation checklist for a new server subproject at `ltml/server`.

## Summary

Build a standalone Go HTTP server that accepts a `multipart/form-data` request, reads LTML from the first part, accepts subsequent uploaded files as request-scoped rendering assets, renders a PDF, streams the PDF response, and cleans up all temporary request artifacts at the end of the request.

## Request And Response Contract

- [ ] Serve a standard `net/http` endpoint at `POST /render`.
- [ ] Require request `Content-Type: multipart/form-data`.
- [ ] Require the first multipart part to be the LTML document.
- [ ] Require the LTML part field name to be `ltml`.
- [ ] Document the preferred LTML part content type as `application/vnd.rowland.leadtype.ltml+xml`.
- [ ] Accept `application/xml`, `text/xml`, or empty LTML part content type as compatibility fallbacks.
- [ ] Accept all following multipart parts as uploaded files for request-scoped rendering.
- [ ] Require uploaded file parts to use field name `file`.
- [ ] Use each uploaded part's multipart `filename` as the virtual asset path.
- [ ] Allow filenames such as `logo.png` and `assets/logo.png`.
- [ ] Reject empty filenames.
- [ ] Reject absolute paths.
- [ ] Reject paths containing `..` or any cleaned path that escapes the request overlay root.
- [ ] Return the rendered PDF with `Content-Type: application/pdf`.
- [ ] Set `Content-Disposition: inline; filename="output.pdf"`.

## Package Layout

- [ ] Create `ltml/server` as an executable `package main`.
- [ ] Add `main.go` for config parsing and server startup.
- [ ] Add `config.go` for runtime configuration definition and validation.
- [ ] Add `handler.go` for multipart parsing, temp-dir lifecycle, and HTTP responses.
- [ ] Add `overlayfs.go` for the request-scoped overlay filesystem.
- [ ] Add `render.go` for LTML parse/render orchestration.
- [ ] Keep the initial implementation in one package unless internal subpackages become clearly useful.

## Startup And Configuration

- [ ] Add dependency on `github.com/namsral/flag`.
- [ ] Parse flags and environment variables at process start using `namsral/flag`.
- [ ] Define `listen` / `LISTEN` with default `:8080`.
- [ ] Define `base-path` / `BASE_PATH` as a required setting for static content.
- [ ] Define `max-upload-bytes` / `MAX_UPLOAD_BYTES` with a reasonable default such as `32<<20`.
- [ ] Define optional `read-timeout` / `READ_TIMEOUT`.
- [ ] Define optional `write-timeout` / `WRITE_TIMEOUT`.
- [ ] Validate that `base-path` exists.
- [ ] Validate that `base-path` is a directory.
- [ ] Construct a standard `http.Server`.
- [ ] Register `POST /render`.
- [ ] Fail fast at startup if configuration is invalid.

## Multipart Handling

- [ ] Wrap the request body with `http.MaxBytesReader`.
- [ ] Parse multipart content as a stream rather than buffering all file content in memory.
- [ ] Enforce that the first part is the LTML part.
- [ ] Return `400 Bad Request` if LTML is missing.
- [ ] Return `400 Bad Request` if LTML is not the first part.
- [ ] Return `400 Bad Request` if the LTML part uses the wrong field name.
- [ ] Read LTML bytes from the first part into memory for parsing.
- [ ] Create a request temp directory with `os.MkdirTemp`.
- [ ] Store uploaded files under the request temp directory using their validated relative filenames.
- [ ] Create parent directories as needed for nested filenames.
- [ ] Keep temp-directory cleanup in a deferred `os.RemoveAll`.

## Overlay Filesystem

- [ ] Implement a request-scoped overlay filesystem using the standard library.
- [ ] Use `os.DirFS(basePath)` as the lower filesystem.
- [ ] Use `os.DirFS(requestUploadDir)` as the upper filesystem.
- [ ] Resolve lookups against the uploaded-content filesystem first.
- [ ] Fall back to the base-path filesystem when the uploaded file is absent.
- [ ] Keep the overlay read-only from the renderer's perspective.
- [ ] Implement `Open`.
- [ ] Implement `ReadFile` if useful for LTML asset consumers.
- [ ] Implement `Stat` if useful for LTML asset consumers.
- [ ] Implement `ReadDir` if useful for LTML asset consumers.
- [ ] Ensure the overlay instance is created per request.
- [ ] Ensure uploaded `logo.png` shadows base-path `logo.png` only for the active request.
- [ ] Ensure parallel requests do not share overlay state.

## LTML Filesystem Threading Sub-Plan

- [ ] Add a request-local asset filesystem seam to `ltml`.
- [ ] Avoid storing request asset state in package-global variables.
- [ ] Stop relying on the shared package-level default scope for mutable request-specific state.
- [ ] Introduce a per-document root scope so concurrent documents can carry different asset filesystems safely.
- [ ] Extend `ltml.Scope` with methods to store and resolve an asset filesystem.
- [ ] Add `SetAssetFS(fs.FS)` on `ltml.Scope`.
- [ ] Add `AssetFS() fs.FS` on `ltml.Scope`.
- [ ] Add a helper such as `OpenAsset(name string)` if it simplifies render-time consumers.
- [ ] Add a convenience method such as `Doc.SetAssetFS(fs.FS)`.
- [ ] Make nested LTML scopes inherit the nearest parent asset filesystem.
- [ ] Ensure external data lookups during rendering go through the scope/document asset seam instead of direct `os.Open`.
- [ ] Keep this change backward compatible for LTML code paths that do not use external assets.

## Render Pipeline

- [ ] Parse LTML from the first multipart part.
- [ ] Attach the request overlay filesystem to the parsed LTML document before rendering.
- [ ] Create the LTML PDF writer.
- [ ] Create a temp PDF file inside the request temp directory.
- [ ] Render the LTML document to that temp PDF file.
- [ ] Close or finalize the PDF output before starting the HTTP stream.
- [ ] Re-open the temp PDF file for response streaming if needed.
- [ ] Stream the finished PDF to the client.
- [ ] Close file handles before request cleanup runs.

## Cleanup And Lifecycle Guarantees

- [ ] Clean up the request temp directory on successful responses.
- [ ] Clean up the request temp directory on parse failures.
- [ ] Clean up the request temp directory on render failures.
- [ ] Clean up the request temp directory on streaming failures.
- [ ] Ensure uploaded files never modify or copy over the configured base path.
- [ ] Ensure request-scoped overrides are not visible to any other request.

## Error Handling

- [ ] Return `405 Method Not Allowed` for non-`POST` requests.
- [ ] Return `400 Bad Request` for malformed multipart requests.
- [ ] Return `400 Bad Request` for invalid upload filenames or illegal relative paths.
- [ ] Return `413 Request Entity Too Large` when the request exceeds the configured size cap.
- [ ] Return `500 Internal Server Error` for temp-file, parse, render, or stream failures.
- [ ] Avoid sending a partial PDF when rendering fails before headers are committed.
- [ ] Log concise failure details that help debug without leaking unnecessary request content.

## Testing Checklist

### Server behavior

- [ ] Add a handler test for a valid LTML-only multipart request returning a PDF.
- [ ] Add a handler test for a valid LTML request with uploaded assets.
- [ ] Add a handler test that rejects a missing LTML part.
- [ ] Add a handler test that rejects LTML when it is not the first part.
- [ ] Add a handler test that rejects invalid multipart filenames.
- [ ] Add a handler test that returns `413` for oversized requests.

### Overlay behavior

- [ ] Add a test proving uploaded files shadow same-named base-path files.
- [ ] Add a test proving base-path files are used when uploads do not provide an override.
- [ ] Add a concurrency test proving different requests do not see each other's uploads.

### LTML filesystem threading

- [ ] Add a test proving the document-specific asset filesystem is request-local.
- [ ] Add a test proving nested scopes inherit the root asset filesystem.
- [ ] Add a test proving one document's asset filesystem does not affect another document.

### Cleanup behavior

- [ ] Add a test proving request temp directories are removed after successful rendering.
- [ ] Add a test proving request temp directories are removed after render failure.

### Repository safety checks

- [ ] Run `go build ./...`.
- [ ] Run `go test ./...`.

## Defaults And Assumptions

- [ ] `ltml/server` is a command, not a reusable public API package.
- [ ] The only initial endpoint is `POST /render`.
- [ ] Uploaded asset paths come from multipart filenames.
- [ ] The preferred private LTML media type is `application/vnd.rowland.leadtype.ltml+xml`.
- [ ] `base-path` is read-only static content configured once at startup.
- [ ] The LTML asset seam belongs in document/scope state, not in the `ltml.Writer` interface.
- [ ] Writing the PDF to a temp file before streaming is intentional and part of the request lifecycle.
