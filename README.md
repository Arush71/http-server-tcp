# 🌐 HTTP/1.1 Server from Scratch in Go

An HTTP/1.1 server implemented directly over raw TCP sockets in Go — **no `net/http`, no frameworks, no abstractions**.

This project focuses on understanding HTTP at the protocol level by manually implementing request parsing, response formatting, and streaming behavior according to RFC 7230.

---

## 🚀 What This Project Demonstrates

* Protocol-level understanding of HTTP/1.1
* Streaming parser design with partial-read handling
* RFC-compliant header parsing and validation
* Chunked transfer encoding and HTTP trailers
* Low-level networking with TCP sockets

---

## ✨ Features

### Request Parsing

* HTTP/1.1 request line parsing (method, target, version)
* Streaming **state-machine parser** handling partial TCP reads
* RFC-compliant header parsing with token validation (`isToken`)
* Case-insensitive headers with duplicate merging
* `Content-Length` body parsing with strict enforcement

---

### Response Handling

* Structured response writer enforcing correct order:

  * status line → headers → body
* Supported status codes: 200, 400, 500
* Chunked transfer encoding (write path)
* HTTP trailers support
* Default headers: `Content-Length`, `Content-Type`, `Connection: close`

---

### Server & Concurrency

* TCP accept loop using `net.Listener`
* One goroutine per connection
* Graceful listener shutdown via signals
* Atomic state handling for shutdown safety

---

### Application Layer

* Reverse proxy: `/httpbin/*` → `https://httpbin.org/*`
* Streaming proxy responses using chunked encoding
* SHA-256 response hash returned as HTTP trailer
* Binary file serving (`/video`)
* Basic route handling via switch-based dispatch

---

## 🏗️ Architecture

### Connection Flow

```
conn.Read() → state-machine parser → request object → handler → response writer → conn.Close()
```

### TCP Handling

* Accept loop runs continuously in a goroutine
* Each connection is handled independently (`go s.handle(conn)`)
* Connections are one-shot (`Connection: close`)

---

### State Machine Parser

```
Initialized → ParsingHeaders → ParsingBody → Done
```

* Incremental parsing based on available bytes
* Handles arbitrary TCP fragmentation correctly
* Consumes and shifts buffer as parsing progresses

---

### Buffer Strategy

* Initial buffer: 1024 bytes
* Automatically doubles when full
* Ensures efficient handling of large or fragmented requests

---

## 🧠 Technical Highlights

### Partial Read Handling

The parser is designed around real TCP behavior — data may arrive in arbitrary chunk sizes.
A cursor-based buffer system ensures correctness across fragmented reads.

---

### RFC-Compliant Header Validation

Header field names are validated against RFC 7230 token rules.
Invalid inputs are rejected instead of silently accepted.

---

### Chunked Encoding + Trailers

Implements HTTP/1.1 chunked transfer encoding, including trailer support for metadata like SHA-256 hashes of streamed responses.

---

### Thoughtful Testing

Uses a custom `chunkReader` to simulate real-world network conditions by limiting read sizes, ensuring the parser behaves correctly under partial reads.

---

## ⚠️ Limitations

This is a **learning-focused implementation**, not a production server.

* No keep-alive (always `Connection: close`)
* No request-side chunked decoding
* Limited status code support
* No TLS / HTTPS
* No connection timeouts
* Basic routing only
* No HTTP/2 or HTTP/3
* No production hardening (rate limiting, pooling, etc.)

---

## 🎯 Why This Project Exists

Most applications rely on high-level abstractions like `net/http`, which hide how HTTP actually works.

This project removes those abstractions to explore:

* how requests are framed over TCP
* how headers and bodies are parsed
* how responses are constructed at the byte level

---

## 🛠️ Getting Started

**Run the server:**

```bash
go run ./cmd/httpserver/main.go
```

Server runs on:

```text
http://localhost:42069
```

---

### Example Requests

```bash
# Default route
curl -v http://localhost:42069/

# Reverse proxy
curl -v --raw http://localhost:42069/httpbin/get

# Video file
curl -v http://localhost:42069/video --output out.mp4
```

---

## 📂 Project Structure

```
cmd/        # entry points and handlers
internal/
  ├── request/   # state machine parser
  ├── headers/   # header parsing + validation
  ├── response/  # response writer + chunked encoding
  └── server/    # TCP loop + connection handling
```

---

## 🧩 What This Shows

* Ability to work below frameworks and abstractions
* Understanding of real-world networking behavior
* Implementation of protocol-level logic from specification
* Careful handling of edge cases and streaming data

---

## 🏁 Final Note

This project prioritizes **depth over completeness** — focusing on how HTTP works internally rather than building a production-ready server.

An HTTP/1.1 server built directly over raw TCP sockets in Go — no `net/http`, no framework, no shortcuts. Every byte of the request is parsed manually according to the HTTP/1.1 specification (RFC 7230).

Built to understand how the protocol actually works at the wire level, not just how to use it.

---

## Features

**Request Parsing**
- Full HTTP/1.1 request line parsing: method (uppercase-only, validated), request target, version
- Streaming, state-machine-based parser that handles partial TCP reads correctly
- Header parsing with RFC 7230 field-name token validation (`isToken`)
- Case-insensitive header storage (keys normalized to lowercase)
- Duplicate header merging via `, ` joining (per spec)
- `Content-Length`-based body accumulation with length enforcement

**Response Writing**
- `response.Writer` abstraction that enforces correct write order: status line → headers → body
- Status codes: 200 OK, 400 Bad Request, 500 Internal Server Error
- Chunked transfer encoding (write path): hex chunk sizes, `\r\n` framing, terminal chunk
- HTTP trailers support
- Default headers: `Content-Length`, `Content-Type`, `Connection: close`

**Server & Concurrency**
- TCP accept loop using `net.Listener`
- One goroutine per accepted connection
- `atomic.Bool` for safe shutdown signaling
- Signal handling (`SIGINT`/`SIGTERM`) for clean listener shutdown

**Application Layer**
- Reverse proxy: `/httpbin/*` → `https://httpbin.org/<path>`, streamed with chunked encoding
- SHA-256 hash of the proxied response body sent as an HTTP trailer (`X-Content-SHA256`)
- Binary file serving: `/video` returns a local MP4 with `Content-Type: video/mp4`
- Hard-coded HTML routes: `/yourproblem` (400), `/myproblem` (500), default (200)

---

## Architecture

### TCP Accept Loop

```
main()
  └── server.Serve(port, handler)
        ├── net.Listen("tcp", ":42069")
        └── go s.listen()
              └── for { conn, _ := listener.Accept(); go s.handle(conn) }
```

The accept loop runs in a background goroutine. Each accepted connection is dispatched to its own goroutine immediately, so the loop is never blocked by connection handling.

### Connection Handling

```
s.handle(conn net.Conn)
  ├── request.RequestFromReader(conn)   // parse the full HTTP request
  ├── response.NewWriter(conn)          // wrap conn in a response writer
  └── s.Handler(rw, req)               // dispatch to application handler
  └── defer conn.Close()
```

### State Machine Parser

The request parser uses four explicit states:

```
Initialized → RequestStateParsingHeaders → ParsingBody → Done
```

`RequestFromReader` reads from the connection into a buffer and repeatedly calls `parse`, which calls `parseSingle` per state transition. `parseSingle` returns the number of bytes it consumed, allowing the outer loop to shift the buffer and continue reading. This correctly handles the reality that a single `conn.Read()` call can return any number of bytes — less than a full line, or spanning multiple headers.

### Buffer Growth Strategy

The read buffer starts at 1024 bytes. When it is full (`cap(buf) - readToIndex == 0`), a new buffer of double the capacity is allocated and the unprocessed data is copied forward. This avoids blocking on large requests while keeping allocation simple.

### Request → Handler → Response Flow

```
1. conn.Read()         → raw bytes into buffer
2. r.parse()           → advance state machine, consume bytes, shift buffer
3. (repeat until Done)
4. Handler(rw, req)    → application logic writes status + headers + body
5. conn.Close()        → connection torn down (Connection: close always)
```

---

## Technical Highlights

### Partial Read Handling
TCP is a byte stream. `conn.Read()` returns however many bytes happen to be available — often less than a full header line or even a full request. The parser is built around this: it tracks a `readToIndex` cursor, shifts consumed bytes out of the buffer after each parse pass, and only advances state when a complete logical unit (request line, header line, body) has been received. The test suite deliberately exercises this using a `chunkReader` that limits reads to N bytes per call.

### RFC-Compliant Header Token Validation
The `isToken()` function in `internal/headers/headers.go` validates header field-names against the RFC 7230 §3.2.6 token definition: visible ASCII only, excluding the set of delimiter characters (`"`, `(`, `)`, `,`, `/`, `:`, `;`, `<`, `=`, `>`, `?`, `@`, `[`, `\`, `]`, `{`, `}`). Invalid field-names return a parse error rather than silently accepting bad input.

### Duplicate Header Merging
When the same field-name appears more than once, values are combined as `"existing, new"` — the correct behavior per RFC 7230 §3.2.2. Header storage is a `map[string]string` with lowercase keys, so `Host`, `host`, and `HOST` all resolve to the same slot.

### Chunked Transfer Encoding + Trailers
The response writer implements the chunked encoding wire format:
```
<hex-length>\r\n
<chunk-data>\r\n
...
0\r\n
<trailer-name>: <trailer-value>\r\n
\r\n
```
The reverse proxy uses this to stream `httpbin.org` responses without buffering the entire body, then appends `X-Content-SHA256` and `X-Content-Length` as HTTP trailers computed over the full streamed body.

### Test Strategy: `chunkReader`
Rather than passing complete request strings to the parser, the request tests use a `chunkReader` — a custom `io.Reader` that returns exactly N bytes per `Read()` call (configurable per test). This directly stress-tests the incremental parsing logic under conditions that match real network behavior.

---

## Limitations / Non-Goals

This is an educational implementation. The following are deliberate non-goals:

| Area | Status |
|---|---|
| Keep-alive / persistent connections | Not implemented — always `Connection: close` |
| Request-side chunked decoding | Not implemented — only `Content-Length` on ingress |
| Status codes | Only 200, 400, 500 |
| Connection timeouts | Not set — `SetReadDeadline`/`SetWriteDeadline` never called |
| TLS / HTTPS | Not implemented |
| Routing | Plain `switch` + one prefix check; no method dispatch, no path params |
| Query string parsing | Not implemented |
| HTTP pipelining | Not implemented |
| HTTP/2 or HTTP/3 | Out of scope |
| Production hardening | Rate limiting, max connections, graceful drain — none present |

The project is also **not** a general-purpose HTTP library. It exists to demonstrate protocol understanding, not to be reused.

---

## Why This Exists

Most Go programs use `net/http` and never see what happens on the wire. This project removes that abstraction to answer the question: *what actually goes over a TCP connection when you make an HTTP request?*

Working through request framing, CRLF delimiters, header token rules, chunked encoding, and trailers at the byte level builds a kind of understanding that reading documentation does not. This project is that exercise.

---

## Getting Started

**Prerequisites:** Go 1.21+

**Run the server:**
```bash
go run ./cmd/httpserver/main.go
```

The server listens on port `42069`.

**Example requests:**

```bash
# Default route
curl -v http://localhost:42069/

# 400 route
curl -v http://localhost:42069/yourproblem

# 500 route
curl -v http://localhost:42069/myproblem

# Reverse proxy with chunked encoding and SHA-256 trailer
curl -v --raw http://localhost:42069/httpbin/get

# Video file (requires assets/vim.mp4 to exist)
curl -v http://localhost:42069/video --output out.mp4
```

**Run tests:**
```bash
go test ./...
```

---

## Project Structure

```
.
├── cmd/
│   ├── httpserver/main.go   # entry point, routing, application handlers
│   ├── tcplistener/         # standalone TCP listener (exploratory)
│   └── udpsender/           # standalone UDP sender (exploratory)
└── internal/
    ├── request/
    │   ├── request.go        # state machine parser, RequestFromReader
    │   └── request_test.go   # partial-read tests via chunkReader
    ├── headers/
    │   ├── headers.go        # header map, Parse, isToken, Set/Get/Delete
    │   └── headers_test.go
    ├── response/
    │   └── response.go       # Writer, WriteStatusLine, chunked TE, trailers
    └── server/
        └── server.go         # TCP accept loop, goroutine dispatch, Handler type
```

---

## What This Demonstrates

- Systems-level thinking: working directly with TCP streams, byte buffers, and wire protocols
- Protocol implementation: following an RFC rather than using an abstraction
- Streaming parser design: state machines, partial reads, buffer management
- Go concurrency primitives: goroutines, `sync/atomic`, signal handling
- Deliberate testing: stress-testing partial reads, not just happy-path strings

