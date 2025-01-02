package main

import (
	"fmt"
	"testing"
)

func Test_Server(t *testing.T) {
	server := NewHTTPServer()
	server.Get("/a/b/c", func(ctx *Context) {
		fmt.Println("handler1")
		_, _ = ctx.ResponseWriter.Write([]byte("hello world"))
	})
	server.Get("/d/*", func(ctx *Context) {
		fmt.Println(ctx.Request.URL.Path)
		fmt.Println("handler2")
		_, _ = ctx.ResponseWriter.Write([]byte("hello world"))
	})

	server.Start(":8081")
}
