package test

import (
	"fmt"
	"github.com/chongyanovo/zweb"
	"testing"
)

func Test_Server(t *testing.T) {
	server := main.NewHTTPServer()
	server.Get("/a/b/c", func(ctx *main.Context) {
		fmt.Println("handler1")
		_, _ = ctx.ResponseWriter.Write([]byte("hello world"))
	})

	server.Start(":8081")
}
