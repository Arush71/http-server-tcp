package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpProtocols/internal/headers"
	"io"
	"strconv"
	"strings"
)

type ParserStateType int

const (
	Initialized ParserStateType = iota
	RequestStateParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserStateType
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case Initialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return bytesConsumed, err
		}
		if bytesConsumed == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.ParserState = RequestStateParsingHeaders
		return bytesConsumed, nil
	case RequestStateParsingHeaders:
		n, isDone, err := r.Headers.Parse(data)
		if err != nil {
			return n, err
		}
		if !isDone {
			return n, nil
		}
		r.ParserState = ParsingBody
		return n, nil
	case ParsingBody:
		contentLength := r.Headers.Get("Content-Length")
		if contentLength == "" {
			r.ParserState = Done
			return 0, nil
		}
		n, err := strconv.Atoi(contentLength)
		if err != nil || n < 0 {
			return 0, errors.New("Invalid Content-Length.")
		}
		r.Body = append(r.Body, data...)

		if len(r.Body) > n {
			return len(data), errors.New("Body length is greater then specified in content-length.")
		}
		if len(r.Body) == n {
			r.ParserState = Done
			return len(data), nil
		}
		return len(data), nil
	case Done:
		return 0, nil
	default:
		return 0, fmt.Errorf("Error: unknown state.")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return n, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

const bufferSize = 1024

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	r := &Request{
		ParserState: Initialized,
		Headers:     make(headers.Headers),
	}
	readToIndex := 0
	for r.ParserState != Done {
		if cap(buf)-readToIndex == 0 {
			newSlice := make([]byte, readToIndex*2)
			copy(newSlice, buf)
			buf = newSlice
		}

		// slice relocation

		read, err := reader.Read(buf[readToIndex:])
		if read > 0 {
			readToIndex += read
			// reading
			bytesConsumed, err := r.parse(buf[:readToIndex])
			if err != nil {
				return nil, err
			}
			copy(buf, buf[bytesConsumed:readToIndex])
			readToIndex -= bytesConsumed
			// parsing
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				if r.ParserState != Done {
					return r, fmt.Errorf("Invalid data")
				}
				break
			}
			return r, fmt.Errorf("Unexpected Error")
		}
	}
	return r, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	before, _, ok := bytes.Cut(data, []byte("\r\n"))
	if !ok {
		return nil, 0, nil
	}
	consumedBytes := len(before) + 2
	Parts := strings.Fields(string(before))
	if len(Parts) != 3 {
		return nil, consumedBytes, errors.New("Request line is not of the right format")
	}
	method := Parts[0]
	for i := range method {
		b := method[i]
		if b < 'A' || b > 'Z' {
			return nil, consumedBytes, errors.New("Request line is not of the right format, methods should be alphabetically capital")
		}
	}
	_, httpVersion, ok := strings.Cut(Parts[2], "/")
	if httpVersion != "1.1" || !ok {
		return nil, consumedBytes, errors.New("Request line is not of the right format, http version error, only supports 1.1")
	}

	return &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: Parts[1],
		Method:        method,
	}, consumedBytes, nil
}
