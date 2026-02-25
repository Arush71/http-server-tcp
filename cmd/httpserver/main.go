package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpProtocols/internal/request"
	"httpProtocols/internal/response"
	"httpProtocols/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	hdr "httpProtocols/internal/headers"
)

const port = 42069

func main() {
	server, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
func myHandler(w *response.Writer, req *request.Request) {
	path := req.RequestLine.RequestTarget

	if after, ok := strings.CutPrefix(path, "/httpbin/"); ok {
		targetPath := after
		targetURL := fmt.Sprintf("https://httpbin.org/%s", targetPath)

		resp, err := http.Get(targetURL)
		if err != nil {
			w.WriteStatusLine(response.BAD_REQUEST)
			_headers := response.GetDefaultHeaders(0)
			_headers.Set("Content-Type", "text/plain")
			_headers.Set("Trailer", "X-Content-SHA256, X-Content-Length")
			w.WriteHeaders(_headers)
			w.WriteBody([]byte("Upstream request failed\n"))
			return
		}
		defer resp.Body.Close()

		// Write status line (always 200 for proxy success)
		w.WriteStatusLine(response.OK)

		// Build chunked headers
		headers := response.GetDefaultHeaders(0)
		headers.Set("Content-Type", resp.Header.Get("Content-Type"))
		headers.Set("Transfer-Encoding", "chunked")
		headers.Set("Trailer", "X-Content-SHA256, X-Content-Length")
		headers.Delete("Content-Length") // make sure length is removed
		w.WriteHeaders(headers)

		// Stream body chunk by chunk
		buffer := make([]byte, 1024)
		var fullBody []byte
		for {

			n, err := resp.Body.Read(buffer)
			if n > 0 {
				w.WriteChunkedBody(buffer[:n])
				fullBody = append(fullBody, buffer[:n]...)
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
		}
		w.WriteChunkedBodyDone()
		hash := sha256.Sum256(fullBody)
		hexStr := hex.EncodeToString(hash[:])
		trailers := hdr.Headers{}
		trailers.Set("X-Content-SHA256", hexStr)
		trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
		w.WriteTrailers(trailers)
		return
	}
	var body []byte
	var status response.StatusCode
	contentType := "text/html"
	switch path {
	case "/yourproblem":
		status = response.BAD_REQUEST
		body = []byte(`<html>
  <head><title>400 Bad Request</title></head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

	case "/myproblem":
		status = response.INTERNAL_SERVER_ERROR
		body = []byte(`<html>
  <head><title>500 Internal Server Error</title></head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
	case "/video":
		data, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			status = response.INTERNAL_SERVER_ERROR
			body = []byte(`<html>
  <head><title>500 Internal Server Error</title></head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

		} else {
			status = response.OK
			body = data
			contentType = "video/mp4"
		}

	default:
		status = response.OK
		body = []byte(`<html>
  <head><title>200 OK</title></head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Set("Content-Type", contentType)

	w.WriteStatusLine(status)
	w.WriteHeaders(headers)
	w.WriteBody(body)
}
