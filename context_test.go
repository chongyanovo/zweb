package main

import (
	"fmt"
	"testing"
)

func Test_Context(t *testing.T) {
	server := NewHTTPServer()
	server.Post("/formValue", func(ctx *Context) {
		name, err := ctx.formValue("name").AsString()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(name)
	})
	server.Post("/queryValue", func(ctx *Context) {
		name, err := ctx.queryValue("name").AsString()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(name)
	})
	server.Post("/pathValue/:name/:id", func(ctx *Context) {
		name, err := ctx.pathValue("name").AsString()
		if err != nil {
			t.Fatal(err)
		}
		id, err := ctx.pathValue("id").AsString()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(name)
		fmt.Println(id)
	})
	server.Post("/json", func(ctx *Context) {

		chongyan := struct {
			Name string `json:"name"`
		}{
			Name: "chongyan",
		}
		ctx.ok().json(chongyan)
		ctx.json(23)
		ctx.json("chongyan")
	})

	server.Start(":9999")
}
