# Plan: HTTP File Upload Support

## Context
Adding file upload handling to `http_server()`. Currently, multipart form parsing reads text fields (`MultipartForm.Value`) but completely ignores `MultipartForm.File`. The uploaded file data is never surfaced to handler scripts.

The goal: when `uploads.enabled = true`, handler scripts can access uploaded files via `req.files`. MIME type determines whether the content is a string (text-based) or binary value (binary/unknown), so apps can directly manipulate text/JSON uploads while efficiently handling images and other binary data.

---

## Config API (Duso-side)

```duso
server = http_server({
  port = 8080,
  uploads = {
    enabled = true,       // required, default false
    max_size = 10240,     // max KB per file (default: 10240 = 10MB)
    timeout = 30          // seconds (reserved, defaults to socket timeout for now)
  }
})
```

## Handler API (Duso-side)

```duso
ctx = context()
req = ctx.request()

// Single file upload
file = req.files.avatar
print(file.filename)              // "avatar.png"
print(file.content_type)          // "image/png"
print(file.size)                  // byte count

// file.data is binary for images, string for text/json/etc.
if type(file.data) == "binary" then
  save_binary(file.data, "/STORE/uploads/" + file.filename)
elseif type(file.data) == "string" then
  parsed = parse_json(file.data)
end

// Multiple files on same field → array
for f in req.files.attachments do
  save_binary(f.data, "/STORE/" + f.filename)
end
```

---

## Text vs Binary MIME Detection

A file is treated as **text** (content = string) if its Content-Type matches:
- `text/*` (text/plain, text/html, text/csv, text/xml, etc.)
- `application/json`
- `application/xml`
- `application/xhtml+xml`
- `application/javascript`
- `application/x-yaml` / `application/yaml`

Everything else → **binary** (content = binary value with metadata).

---

## Files to Modify

### 1. `pkg/runtime/http_server.go`

**Add `UploadConfig` struct** (after `JWTConfig` ~line 41):
```go
type UploadConfig struct {
    Enabled bool
    MaxSize int64         // max bytes per file (converted from KB at parse time)
    Timeout time.Duration // reserved for future dedicated upload timeout
}
```

**Add `Upload UploadConfig` field** to `HTTPServerValue` (after `JWT JWTConfig` ~line 57).

**Add `isTextMIME(contentType string) bool` helper function**:
- Returns true for text/*, application/json, application/xml, application/xhtml+xml, application/javascript, application/yaml, application/x-yaml

**Update `handleRequest`** (~line 854): pass `Upload UploadConfig` into `RequestContext` creation.

**Update `GetRequest()` multipart branch** (~line 1235):
- After parsing text form fields from `MultipartForm.Value`, if `rc.Upload.Enabled`:
  - Enforce max size using `http.MaxBytesReader` (already called before `ParseMultipartForm`)
  - Iterate `rc.Request.MultipartForm.File`
  - For each `[]*multipart.FileHeader`:
    - Open the file, read bytes
    - Check size against `rc.Upload.MaxSize` (skip / error if exceeded)
    - Detect MIME: use `header.Header.Get("Content-Type")`, fallback to `mime.TypeByExtension(filepath.Ext(header.Filename))`
    - If `isTextMIME`: content = `string(bytes)`; wrap in object `{data, filename, content_type, size}`
    - If binary: call `script.NewBinary(bytes)`, set `Metadata["filename"]`, `Metadata["content_type"]`, `Metadata["size"]`; wrap in object same shape
    - Handle multiple files per field: collect into `[]any` slice
  - Build `filesMap map[string]any`

**Add `"files": filesMap` to result map** in `GetRequest()` (~line 1267).
- Always include key (empty map if uploads disabled or no files), so `req.files` is never nil.

### 2. `pkg/runtime/goroutine_context.go`

**Add `Upload UploadConfig` field** to `RequestContext` struct (after `CacheControl string`):
```go
Upload UploadConfig
```

### 3. `pkg/runtime/builtin_http_server.go`

**Add uploads config parsing** (after JWT parsing block ~line 253):
```go
// Parse uploads config
if uploadRaw, ok := config["uploads"]; ok {
    if uploadMap, ok := uploadRaw.(map[string]any); ok {
        if enabled, ok := uploadMap["enabled"].(bool); ok {
            server.Upload.Enabled = enabled
        }
        if maxSizeKB, ok := uploadMap["max_size"].(float64); ok {
            server.Upload.MaxSize = int64(maxSizeKB) * 1024 // convert KB → bytes
        }
        if timeout, ok := uploadMap["timeout"].(float64); ok {
            server.Upload.Timeout = time.Duration(timeout) * time.Second
        }
    }
}
```

Default MaxSize: `10 * 1024 * 1024` bytes (10MB = 10240KB) set when `HTTPServerValue` is initialized.

### 4. `pkg/runtime/http_server.go` — `ParseMultipartForm` size limit

Update the hardcoded `32 << 20` (32MB) in the multipart call to use `rc.Upload.MaxSize` when uploads are enabled, or a safe default when disabled.

### 5. `docs/reference/binary.md`

Update the "HTTP File Uploads" section to reflect the new object-based API (replacing the old pattern that showed `req.files["avatar"]` as a raw binary value with bracket metadata).

### 6. `docs/reference/http_server.md`

Add a new **File Uploads** section documenting:
- The `uploads` config object and its fields
- `req.files` access pattern with both text and binary examples
- MIME type routing logic

---

## Verification

1. Build: `./build.sh`
2. Write a test script with `http_server` that accepts a POST with a file upload
3. Test with curl:
   ```bash
   # Binary (image)
   curl -X POST http://localhost:8080/upload -F "avatar=@test.png"
   # Text (JSON)
   curl -X POST http://localhost:8080/upload -F "data=@config.json"
   ```
4. Confirm `req.files.avatar.data` is binary type, `req.files.data.data` is string
5. Confirm `file.filename`, `file.content_type`, `file.size` fields are correct
6. Test max_size enforcement: upload a file exceeding the limit, expect 413 or error
