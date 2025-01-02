package main

import (
	"fmt"
	"regexp"
	"strings"
)

type nodeType uint8

const (
	nodeTypeStatic nodeType = iota
	nodeTypeRoot
	nodeTypeParam
	nodeTypeRegexp
	nodeTypeWild
)

// router 路由
type router struct {
	trees map[string]*node
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}

// node 路由节点
type node struct {
	nType       nodeType
	path        string
	children    map[string]*node
	wildChild   *node
	paramChild  *node
	regexpChild *node
	regexpExpr  *regexp.Regexp
	paramName   string
	handler     HandleFunc
}

// newRouter 创建路由
func newRouter() *router {
	return &router{
		trees: make(map[string]*node),
	}
}

// addRoute 添加路由节点
// 对于已经注册的路由,无法被覆盖
// path必须以 / 开头,不能以 / 结尾
// 不能在同一个位置注册不同的参数路由
// 不能在同一个位置注册不同的参数路由和通配符路由
// 同名路径参数在路由匹配时会被覆盖
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
	if len(path) != 1 && path[len(path)-1] == '/' {
		panic(fmt.Sprintf("path must not end with '/': %s", path))
	}
	root, ok := r.trees[method]
	if !ok {
		root = &node{
			nType: nodeTypeRoot,
			path:  "/",
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
	lastNode := root
	for _, p := range strings.Split(path[1:], "/") {
		if p == "" {
			panic(fmt.Sprintf("path must not contain empty path: %s", path))
		}
		lastNode = lastNode.childOrCreate(p)
	}
	if lastNode.handler != nil {
		panic(fmt.Sprintf("have the same route: %s", path))
	}
	lastNode.handler = handleFunc
}

// childOrCreate 创建子节点
func (n *node) childOrCreate(path string) *node {
	if path[0] == ':' {
		paramName, expr, isRegexp := n.parseParam(path)
		if isRegexp {
			return n.childOrCreateRegexp(path, expr, paramName)
		}
		return n.childOrCreateParam(path, paramName)
	}
	if path == "*" {
		return n.childOrCreateWild(path)
	}
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{
			nType: nodeTypeStatic,
			path:  path,
		}
		n.children[path] = child
	}
	return child
}

// parseParam 解析参数路由
func (n *node) parseParam(path string) (string, string, bool) {
	path = path[1:]
	regs := strings.SplitN(path, "(", 2)
	if len(regs) == 2 {
		expr := regs[1]
		if strings.HasSuffix(expr, ")") {
			return regs[0], expr[:len(expr)-1], true
		}
	}
	return path, "", false
}

func (n *node) childOrCreateParam(path string, paramName string) *node {
	if n.regexpChild != nil {
		panic(fmt.Sprintf("exist regular routes. parameters routes cannot be registered: %s", path))
	}
	if n.wildChild != nil {
		panic(fmt.Sprintf("exist wildcard routes. parameters routes cannot be registered: %s", path))
	}
	if n.paramChild == nil {
		n.paramChild = &node{
			nType: nodeTypeParam,
			path:  path,
		}
	}
	return n.paramChild
}

// childOrCreateRegexp 创建正则表达式子节点
func (n *node) childOrCreateRegexp(path string, expr string, paramName string) *node {
	if n.wildChild != nil {
		panic(fmt.Sprintf("exist wildcard routes. regular routes cannot be registered: %s", path))
	}
	if n.paramChild != nil {
		panic(fmt.Sprintf("exist parameters routes. regular routes cannot be registered: %s", path))
	}
	if n.regexpChild != nil {
		if n.regexpChild.regexpExpr.String() != expr || n.paramName != paramName {
			panic(fmt.Sprintf("routes conflict. route %s already exists. and new route %s fails to be registered.", n.regexpExpr, path))
		}
	} else {
		regexpExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("regexp %w is invalid", err))
		}
		n.regexpChild = &node{
			nType:      nodeTypeRegexp,
			path:       path,
			regexpExpr: regexpExpr,
			paramName:  paramName,
		}
	}
	return n.regexpChild
}

// childOrCreateWild 创建通配符子节点
func (n *node) childOrCreateWild(path string) *node {
	if n.paramChild != nil {
		panic(fmt.Sprintf("exist parameters routes. wildcards routes cannot be registered: %s", path))
	}
	if n.regexpChild != nil {
		panic(fmt.Sprintf("exist regular routes. wildcards routes cannot be registered: %s", path))
	}
	if n.wildChild == nil {
		n.wildChild = &node{
			nType: nodeTypeWild,
			path:  path,
		}
	}
	return n.wildChild
}

// findRouter 查找路由节点
func (r *router) findRouter(method, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return &matchInfo{
			n: root,
		}, true
	}
	mi := &matchInfo{}
	for _, p := range strings.Split(strings.Trim(path, "/"), "/") {
		var child *node
		child, ok = root.childOf(p)
		if !ok {
			if root.nType == nodeTypeWild {
				mi.n = root
				return mi, true
			}
			return nil, false
		}
		if child.paramName != "" {
			mi.addValue(child.paramName, p)
		}
		root = child
	}
	mi.n = root
	return mi, true
}

// childOf 查找子节点,优先静态匹配,然后再通配符匹配
// *node 命中子节点
// nodeType 路由类型
// bool 是否命中
func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.childOfNonStatic(path)
	}
	res, ok := n.children[path]
	if !ok {
		return n.childOfNonStatic(path)
	}
	return res, ok
}

// childOfNonStatic 查找非静态匹配子节点
func (n *node) childOfNonStatic(path string) (*node, bool) {
	if n.regexpChild != nil {
		if n.regexpChild.regexpExpr.Match([]byte(path)) {
			return n.regexpChild, true
		}
	}
	if n.paramChild != nil {
		return n.paramChild, true
	}
	return n.wildChild, n.wildChild != nil
}
