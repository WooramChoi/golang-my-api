package server

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	address string
	// logger  *log.Logger
	mux     *http.ServeMux
	Context map[string]interface{}
}

type RouteHandler func(http.ResponseWriter, *http.Request)

type Router interface {
	GetRouter() map[string]RouteHandler
}

func New(address string) *Server {
	server := Server{}
	server.address = address
	server.mux = http.NewServeMux()
	server.Context = make(map[string]interface{})
	// server.Context["logger"] = configLog()
	return &server
}

func loggerMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("[%s] %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (server *Server) AddRoute(route string, handler RouteHandler) {
	server.mux.HandleFunc(route, handler)
}

func (server *Server) AddRouteByRouter(router Router) {
	for route, handler := range router.GetRouter() {
		server.mux.HandleFunc(route, handler)
	}
}

func (server *Server) Run() {
	serverLogger, ok := server.Context["logger"].(*log.Logger)
	if ok {
		loggerMux := loggerMiddleware(server.mux, serverLogger)
		serverLogger.Println("Application Run with logger")
		http.ListenAndServe(server.address, loggerMux)
	} else {
		fmt.Println("Application Run without logger")
		http.ListenAndServe(server.address, server.mux)
	}
}
