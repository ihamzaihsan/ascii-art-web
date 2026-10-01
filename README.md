# ASCII Art Web

A Go web application that turns text into eight-row ASCII artwork. Choose **Standard**, **Shadow**, or **Thinkertoy**, then select and copy the result into a terminal, text file, or message using a monospace font.

Built with Go's standard library, server-rendered HTML, and responsive CSS. There are no external Go modules, JavaScript requirements, database, or frontend build tools.

## Features

- Three bundled banners covering all 95 printable ASCII characters.
- Multiline input, preserved spaces and blank lines, and support for LF and browser CRLF line endings.
- Keyboard-accessible native form controls, visible focus indicators, and a scrollable, read-only output area.
- Server-side input validation: up to 100 characters after CRLF normalization; printable ASCII and newlines only. Newline-only input is rejected; spaces are supported.
- Correct HTTP error responses for invalid input, missing pages, unsupported methods and media types, and oversized forms.
- Banners, templates, CSS, and favicon embedded in a portable executable. Banner data and templates are loaded once at startup.
- Request size limits, HTTP timeouts, automatic HTML escaping, a Content Security Policy, and graceful shutdown on Ctrl+C or SIGTERM.

## Quick start

Prerequisite: **Go 1.23 or newer**. Local verification used Go 1.23.4 on Windows. A C compiler is needed only for the optional race detector; normal build and test commands need no C toolchain.

Clone this repository using its GitHub URL, then enter the project directory:

```sh
cd ascii-art-web
go run .
```

Open **http://127.0.0.1:8080**. Enter `Hello, world!`, choose a style, and select **Generate ASCII art**. Select and copy the result. Use the browser's Back button to revise an earlier submission or **Generate another** for a fresh form.

If the default port is occupied or reserved by your operating system, choose another port with `-addr` and use that port in the browser URL.

Stop with **Ctrl+C**. No environment variables or configuration files are required. The optional `-addr` flag controls the bind address:

```sh
go run . -addr 127.0.0.1:9090
```

For access from another machine, use `-addr 0.0.0.0:8080` and configure the host firewall as appropriate. The server provides plain HTTP; a public deployment needs an HTTPS reverse proxy.

### Build a standalone executable

```sh
go build -o ascii-art-web .
./ascii-art-web -addr 127.0.0.1:8080
```

On Windows:

```powershell
go build -o ascii-art-web.exe .
.\ascii-art-web.exe -addr 127.0.0.1:8080
```

The built executable works from any directory without separate resource files. Rebuild after editing banners, templates, or assets.

### HTTP example

`GET /` serves the form; `POST /ascii-art` returns an HTML result page. It expects URL-encoded fields named `input-text` and `banner`. This is an HTML application, not a JSON API.

```sh
curl --data-urlencode 'input-text=Hello!' --data-urlencode 'banner=shadow' http://127.0.0.1:8080/ascii-art
```

On Windows PowerShell, use `curl.exe` to avoid the legacy `curl` alias. A multiline value can be sent as `--data-urlencode 'input-text@message.txt'` where `message.txt` contains printable ASCII text and line breaks.

| Status | Meaning |
| --- | --- |
| 200 | Form, generated result, or existing asset |
| 400 | Invalid text, unknown banner, or malformed form |
| 404 | Unknown page or asset |
| 405 | Unsupported method; page handlers include an `Allow` header |
| 413 | Form body larger than 4 KiB |
| 415 | POST content type other than `application/x-www-form-urlencoded` |
| 500 | Template rendering failure |

## Checks

Run from the repository root:

```sh
go test ./...
go vet ./...
go test -race ./...
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

For PowerShell, quote arguments containing `=` and file extensions where needed:

```powershell
go test '-coverpkg=./...' '-coverprofile=coverage.out' ./...
go tool cover '-func=coverage.out'
```

Check formatting with `gofmt -l main.go src` (no output means all Go files are formatted).

Tests cover exact glyph rendering, character bounds, all styles and printable characters, newline preservation, and malformed/missing banner data. HTTP handlers, concurrent requests, and process startup/shutdown are not covered by the remaining automated suite. Earlier local checks passed on Windows with Go 1.23.4: tests, race detector, `go vet`, formatting, standalone build, and HTTP smoke tests from a different working directory. `go run` was also verified with `-addr 127.0.0.1:18081`; Windows rejected binding the default port 8080 on the verification machine. GitHub Actions runs formatting, static analysis, race-enabled tests, and a build on Linux. The hosted workflow must run on GitHub to verify that environment.

## Technical design

```text
main.go                     Embedded resources, listen address, timeouts, shutdown
src/asciiart/               Banner validation and pure in-memory rendering
src/server/                 HTTP handlers, request validation, buffered templates
banners/                    Standard, Shadow, and Thinkertoy glyph data
templates/                  Input, result, and error pages
assets/                     Responsive stylesheet and favicon
.github/workflows/ci.yml    Automated Go checks
```

Each banner contains 95 blocks: one empty separator followed by eight glyph rows. Startup validates the block count and separators, then stores glyphs in fixed-size arrays. Rendering joins each character's corresponding row using `strings.Builder`, avoiding file I/O during requests. Adding a style requires updating the renderer's style list, form options, banner data, and tests.

A dedicated `http.ServeMux` routes requests. Validation occurs before rendering; templates execute into a buffer before the status is written, preventing partial successful responses when execution fails. `html/template` escapes user text. Immutable banner and template data are shared across requests.

## Scope and limitations

- Supports printable ASCII (32–126) and newlines; tabs, lone carriage returns, accented letters, emoji, and other Unicode are rejected.
- Output is intended for monospace display and may require horizontal scrolling. Input is limited to 100 characters; there is no file upload or automatic clipboard/download feature.
- No persistence, accounts, or public demo deployment is included. Browser input still needs server validation; this application enforces both.
- Automated tests cover the renderer only; HTTP behavior, generated HTML, browser appearance, and assistive-technology behavior require separate verification.
- No license file is included in the repository.

## Credits

Originally created as a collaborative learning project by **Hamza Cheema** (`hcheema`), **Husain Hanoon** (`hhanoon`), and **Zainab AlNabhan** (`znabhan`). Original contributor credit is retained. This repository includes a subsequent engineering refresh of the implementation, tests, documentation, and interface.
