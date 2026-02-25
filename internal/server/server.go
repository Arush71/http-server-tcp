package server

import (
	"httpProtocols/internal/request"
	"httpProtocols/internal/response"
	"io"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Closed   atomic.Bool
	Handler  Handler
}

func Serve(port int, Handler Handler) (*Server, error) {

	listner, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		Listener: listner,
		Closed:   atomic.Bool{},
		Handler:  Handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	return s.Listener.Close()
}
func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Closed.Load() {
				return
			}
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil || req == nil {
		return
	}

	rw := response.NewWriter(conn)

	s.Handler(rw, req)
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}
type Handler func(w *response.Writer, req *request.Request)

func writeHandlerError(w io.Writer, hErr *HandlerError) error {
	if err := response.WriteStatusLine(w, hErr.StatusCode); err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(hErr.Message))
	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}

	if _, err := w.Write([]byte(hErr.Message)); err != nil {
		return err
	}

	return nil
}
