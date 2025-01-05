package main

import (
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	Request          *http.Request
	ResponseWriter   http.ResponseWriter
	PathParams       map[string]string
	cacheQueryValues url.Values
}

func (ctx *Context) ok() *Context {
	ctx.ResponseWriter.WriteHeader(http.StatusOK)
	return ctx
}

func (ctx *Context) fail(code int) *Context {
	ctx.ResponseWriter.WriteHeader(code)
	return ctx
}

func (ctx *Context) notFound() *Context {
	ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
	_, err := ctx.ResponseWriter.Write([]byte("404 not found"))
	if err != nil {
		panic("write 404 fail")
		return nil
	}
	return ctx
}

func (ctx *Context) cookie(cookie *http.Cookie) *Context {
	http.SetCookie(ctx.ResponseWriter, cookie)
	return ctx
}

func (ctx *Context) json(value any) error {
	ctx.ResponseWriter.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(ctx.ResponseWriter).Encode(value)
}

// bindJson 解析json
func (ctx *Context) bindJson(value any) error {
	if ctx.Request.Body == nil {
		return errors.New("BindJson value is nil")
	}
	decoder := json.NewDecoder(ctx.Request.Body)
	return decoder.Decode(value)
}

// bindYaml 解析yaml
func (ctx *Context) bindYaml(value any) error {
	if ctx.Request.Body == nil {
		return errors.New("BindYaml value is nil")
	}
	decoder := yaml.NewDecoder(ctx.Request.Body)
	return decoder.Decode(value)
}

// formValue 获取表单数据
func (ctx *Context) formValue(key string) StringValue {
	err := ctx.Request.ParseForm()
	if err != nil {
		return StringValue{"", err}
	}
	values, ok := ctx.Request.Form[key]
	if !ok {
		return StringValue{"", errors.New("key not found")}
	}
	return StringValue{values[0], nil}
}

// queryValue 获取查询参数
func (ctx *Context) queryValue(key string) StringValue {
	if ctx.cacheQueryValues == nil {
		ctx.cacheQueryValues = ctx.Request.URL.Query()
	}
	values, ok := ctx.cacheQueryValues[key]
	if !ok {
		return StringValue{"", errors.New("key not found")}
	}
	return StringValue{values[0], nil}
}

// pathValue 获取路径参数
func (ctx *Context) pathValue(key string) StringValue {
	value, ok := ctx.PathParams[key]
	if !ok {
		return StringValue{"", errors.New("key not found")}
	}
	return StringValue{value, nil}
}

type StringValue struct {
	value string
	err   error
}

func (v StringValue) AsString() (string, error) {
	if v.err != nil {
		return "", v.err
	}
	return v.value, v.err
}

func (v *StringValue) AsInt64() (int64, error) {
	if v.err != nil {
		return 0, v.err
	}
	return strconv.ParseInt(v.value, 10, 64)
}
