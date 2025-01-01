package main

import (
	"net"
	"net/http"
)

var _ Server = &HttpServer{}

type Server interface {
	http.Handler
	Start(address string) error
	addRoute(method string, pattern string, handleFunc HandleFunc)
}

type HttpServer struct {
	*router
}

func NewHTTPServer() *HttpServer {
	return &HttpServer{
		router: newRouter(),
	}
}

func (s *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Request:        request,
		ResponseWriter: writer,
	}
	s.serve(ctx)
}

func (s *HttpServer) serve(ctx *Context) {
	n, found := s.router.FindRouter(ctx.Request.Method, ctx.Request.URL.Path)
	if !found || n == nil || n.handler == nil {
		ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		_, _ = ctx.ResponseWriter.Write([]byte("404 not found"))
		return
	}
	n.handler(ctx)
}

func (s *HttpServer) Start(address string) error {
	listen, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	return http.Serve(listen, s)
}

func (s *HttpServer) Router(method string, pattern string, handleFunc HandleFunc) {
	s.addRoute(method, pattern, handleFunc)
}

func (s *HttpServer) Get(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodGet, pattern, handleFunc)
}

func (s *HttpServer) Head(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodHead, pattern, handleFunc)
}

func (s *HttpServer) Post(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPost, pattern, handleFunc)
}

func (s *HttpServer) Put(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPut, pattern, handleFunc)
}

func (s *HttpServer) Patch(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPatch, pattern, handleFunc)
}

func (s *HttpServer) Delete(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodDelete, pattern, handleFunc)
}

func (s *HttpServer) Connect(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodConnect, pattern, handleFunc)
}

func (s *HttpServer) Options(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodOptions, pattern, handleFunc)
}

func (s *HttpServer) Trace(pattern string, handleFunc HandleFunc) {
	s.addRoute(http.MethodTrace, pattern, handleFunc)
}
