# From TCP to HTTP – Minimal HTTP/1.1 Server in Go

A **from-scratch** HTTP/1.1 server implementation built directly on raw TCP sockets — **without** using Go's `net/http` package for request serving.

The main goal of this project is to deeply understand how the HTTP/1.1 protocol actually works under the hood by manually implementing core parts of the specification.

## Features

- TCP-based server using only `net.Listener`
- Manual parsing of raw HTTP/1.1 request byte streams
- Strict CRLF (`\r\n`) line ending handling
- Incremental/state-machine-based request parser (handles partial TCP reads)
- Proper `\r\n\r\n` header termination detection
- Strict `Content-Length` validation & enforcement
- Case-insensitive header storage with support for repeated headers
- Custom HTTP response writer abstraction with write-order validation
- Full **chunked transfer encoding** support
  - Hexadecimal chunk sizes
  - Proper `\r\n` framing
  - Final `0\r\n\r\n` terminator
- HTTP trailers support (used in proxy responses)
- Reverse proxy endpoint `/httpbin/*` → https://httpbin.org/
  - Streaming upstream → downstream forwarding
  - Automatic `Content-Length` removal
  - `Transfer-Encoding: chunked` when streaming
- Binary file serving endpoint `/video`
  - Serves a local MP4 file with correct `Content-Type: video/mp4`
- SHA256 hashing of proxied response body sent via trailers:
  - `X-Content-SHA256`
  - `X-Content-Length`

## Project Structure

├── internal/
│   ├── request/       → incremental HTTP request parser & state machine
│   ├── headers/       → header parsing, storage, case-insensitive map
│   ├── response/      → custom ResponseWriter, chunked encoding, trailers
│   └── server/        → TCP listener loop, connection handling, handler abstraction
├── main.go            → routing table + example handlers
├── go.mod
└── README.md

## Current Limitations / Not Implemented

This is **not** a production server, it's an educational implementation.

Intentionally missing / not implemented:

- TLS / HTTPS
- Keep-Alive & connection pooling optimizations
- HTTP/1.1 pipelining
- HTTP/2 or HTTP/3
- Most security headers & protections
- Request body streaming for very large uploads
- Comprehensive error recovery & timeout handling
- Configurable routing / middleware
- Graceful shutdown

## Why Build This?

Modern frameworks abstract away most of the protocol-level details. I wanted to remove that abstraction and see what actually happens over a TCP connection.

This project was mainly about understanding how HTTP is framed, parsed, and streamed — not just how to use it.

## How to run?

Assuming you have some sort of video file in the assets directory(not required but needed to test binary data implementation):

Navigate to cmd/httpserver and run: "go run main.go".
That's it.

Now you can start making requests(see the handler on the same file for more info).

