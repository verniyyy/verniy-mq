package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/verniyyy/verniy-mq/src"
)

// NewHTTPServer ...
func NewHTTPServer(host string, port int, mqm src.MQManager) Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	h := newMQManagerHandler(mqm)

	r.Route("/api/v1/vmq", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Delete("/{queueName}", h.Delete)
	})

	return httpServer{
		router: r,
		host:   host,
		port:   fmt.Sprint(port),
	}
}

// httpServer ...
type httpServer struct {
	router *chi.Mux
	host   string
	port   string
}

// Run ...
func (s httpServer) Run() {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Printf("Error: %v", err)
	}
}
