package response

import (
	"errors"
	"fmt"
	"httpProtocols/internal/headers"
	"io"
)

type StatusCode int

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case OK:
		w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return nil
	case BAD_REQUEST:
		w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return nil
	case INTERNAL_SERVER_ERROR:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return nil
	default:
		return errors.New("Unknown Error.")
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.Headers{}
	header.Set("Content-Length", fmt.Sprint(contentLen))
	header.Set("Connection", "close")
	header.Set("Content-Type", "text/plain")
	return header
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		fmt.Fprintf(w, "%s: %s\r\n", key, value)
	}
	w.Write([]byte("\r\n"))
	return nil
}

type Writer struct {
	w           io.Writer
	wroteStatus bool
	wroteHeader bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (rw *Writer) WriteStatusLine(statusCode StatusCode) error {
	if rw.wroteStatus {
		return errors.New("status line already written")
	}
	rw.wroteStatus = true
	return WriteStatusLine(rw.w, statusCode)
}

func (rw *Writer) WriteHeaders(headers headers.Headers) error {
	if !rw.wroteStatus {
		return errors.New("must write status line before headers")
	}
	if rw.wroteHeader {
		return errors.New("headers already written")
	}
	rw.wroteHeader = true
	return WriteHeaders(rw.w, headers)
}

func (rw *Writer) WriteBody(p []byte) (int, error) {
	if !rw.wroteHeader {
		return 0, errors.New("must write headers before body")
	}
	return rw.w.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	hStr := fmt.Sprintf("%x", len(p))
	if _, err := w.w.Write([]byte(hStr)); err != nil {
		return 0, err
	}

	// Write CRLF
	if _, err := w.w.Write([]byte("\r\n")); err != nil {
		return 0, err
	}

	// Write actual data
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}

	// Write trailing CRLF
	if _, err := w.w.Write([]byte("\r\n")); err != nil {
		return n, err
	}
	return n, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.w.Write([]byte("0\r\n"))
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		if _, ok := fmt.Fprintf(w.w, "%s: %s\r\n", k, v); ok != nil {
			return ok
		}
	}

	if _, ok := w.w.Write([]byte("\r\n")); ok != nil {
		return ok
	}
	return nil
}
