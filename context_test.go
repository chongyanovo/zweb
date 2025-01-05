package main

import (
	"fmt"
	"testing"
)

func Test_Context(t *testing.T) {
	server := NewHTTPServer()
	server.Post("/formValue", func(ctx *Context) {
		name, err := ctx.FormValue("name").AsString()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(name)
	})
	server.Post("/queryValue", func(ctx *Context) {
		name, err := ctx.QueryValue("name").AsString()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(name)
	})
	server.Post("/pathValue/:name/:id", func(ctx *Context) {
		name, err := ctx.PathValue("name").AsString()
		if err != nil {
			t.Fatal(err)
		}
		id, err := ctx.PathValue("id").AsString()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(name)
		fmt.Println(id)
	})

	server.Start(":9999")
}
