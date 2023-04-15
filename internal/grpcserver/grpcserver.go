// Package grpcserver implements handy wrap around gRPC server
// to group common settings and tasks inside single entity.
package grpcserver

import (
	"net"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"google.golang.org/grpc"
)

// Server wraps gRPC server entity and handy means
// to simplify work with the entity.
type Server struct {
	address entity.NetAddress
	server  *grpc.Server
	notify  chan error
}

// New creates new instance of gRPC server.
func New(address entity.NetAddress) *Server {
	grpcServer := grpc.NewServer()

	s := &Server{
		address: address,
		server:  grpcServer,
		notify:  make(chan error, 1),
	}

	return s
}

// Instance grants access to the underlying gRPC server.
// Should be used to attach new API services.
func (s *Server) Instance() *grpc.Server {
	return s.server
}

// Start launches the gRPC server.
func (s *Server) Start() {
	go func() {
		listen, err := net.Listen("tcp", s.address.String())
		if err != nil {
			s.notify <- err
			return
		}

		s.notify <- s.server.Serve(listen)
		close(s.notify)
	}()
}

// Notify reports errors received during start and work of the server.
// Usually such errors are not recoverable.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	if s.server == nil {
		return
	}

	s.server.GracefulStop()
}
