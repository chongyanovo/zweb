package main

import (
	"fmt"
	"strings"
)

// router 路由
type router struct {
	trees map[string]*node
}

// node 路由节点
type node struct {
	path     string
	children map[string]*node
	handler  HandleFunc
}

// newRouter 创建路由
func newRouter() *router {
	return &router{
		trees: make(map[string]*node),
	}
}

// AddRoute 添加路由节点
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if handleFunc == nil {
		panic(fmt.Sprintf("handleFunc can not be nil: %s", path))
	}
	if path == "" {
		panic(fmt.Sprintf("path can not be empty: %s", path))
	}
	if path[0] != '/' {
		panic(fmt.Sprintf("path must start with '/': %s", path))
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic(fmt.Sprintf("path must not end with '/': %s", path))
	}
	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic(fmt.Sprintf("have the same route: %s", path))
		}
		root.handler = handleFunc
		return
	}
	for _, path := range strings.Split(path[1:], "/") {
		if path == "" {
			panic(fmt.Sprintf("path must not contain empty path: %s", path))
		}
		children := root.childOrCreate(path)
		root = children
	}
	if root.handler != nil {
		panic(fmt.Sprintf("have the same route: %s", path))
	}
	root.handler = handleFunc
}

// childOrCreate 创建子节点
func (n *node) childOrCreate(path string) *node {
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{
			path: path,
		}
		n.children[path] = child
	}
	return child
}

func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return nil, false
	}
	return n.children[path], true
}

// FindRouter 查找路由节点
func (r *router) FindRouter(method, path string) (*node, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	path = strings.Trim(path, "/")
	for _, path := range strings.Split(path, "/") {
		child, found := root.childOf(path)
		if !found {
			return nil, false
		}
		root = child
	}
	return root, true
}
