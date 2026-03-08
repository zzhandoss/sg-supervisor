package releasepanelhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"sg-supervisor/internal/releasepanel"
)

type Server struct {
	listen  string
	service *releasepanel.Service
}

func NewServer(listen string, service *releasepanel.Service) *Server {
	return &Server{listen: listen, service: service}
}

func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              s.listen,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
