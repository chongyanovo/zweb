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

// BindJson 解析json
func (ctx *Context) BindJson(value any) error {
	if ctx.Request.Body == nil {
		return errors.New("BindJson value is nil")
	}
	decoder := json.NewDecoder(ctx.Request.Body)
	return decoder.Decode(value)
}

// BindYaml 解析yaml
func (ctx *Context) BindYaml(value any) error {
	if ctx.Request.Body == nil {
		return errors.New("BindYaml value is nil")
	}
	decoder := yaml.NewDecoder(ctx.Request.Body)
	return decoder.Decode(value)
}

// FormValue 获取表单数据
func (ctx *Context) FormValue(key string) StringValue {
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

// QueryValue 获取查询参数
func (ctx *Context) QueryValue(key string) StringValue {
	if ctx.cacheQueryValues == nil {
		ctx.cacheQueryValues = ctx.Request.URL.Query()
	}
	values, ok := ctx.cacheQueryValues[key]
	if !ok {
		return StringValue{"", errors.New("key not found")}
	}
	return StringValue{values[0], nil}
}

func (ctx *Context) PathValue(key string) StringValue {
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
