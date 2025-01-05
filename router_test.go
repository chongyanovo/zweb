package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_addRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/*/*",
		},
		{
			method: http.MethodPost,
			path:   "/*/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/a/:b/:c",
		},
	}
	mockHandler := func(c *Context) {

	}
	r := newRouter()
	for _, test := range testRoutes {
		r.addRoute(test.method, test.path, mockHandler)
	}
	wantRouter := &router{trees: map[string]*node{
		http.MethodGet: {
			path:  "/",
			nType: nodeTypeRoot,
			children: map[string]*node{
				"user": {
					path:  "user",
					nType: nodeTypeStatic,
					children: map[string]*node{
						"home": {
							path:     "home",
							children: map[string]*node{},
							handler:  mockHandler,
						},
					},
				},
				"order": {
					path:  "order",
					nType: nodeTypeStatic,
					children: map[string]*node{
						"detail": {
							path:     "detail",
							children: map[string]*node{},
							handler:  mockHandler,
						},
					},
				},
			},
			handler: mockHandler,
		},
		http.MethodPost: {
			path:  "/",
			nType: nodeTypeRoot,
			wildChild: &node{
				path:  "*",
				nType: nodeTypeWild,
				wildChild: &node{
					path:    "*",
					nType:   nodeTypeWild,
					handler: mockHandler,
				},
				children: map[string]*node{
					"order": {
						path:  "order",
						nType: nodeTypeStatic,
						wildChild: &node{
							path:    "*",
							nType:   nodeTypeWild,
							handler: mockHandler,
						},
					},
				},
			},
			children: map[string]*node{
				"order": {
					path:  "order",
					nType: nodeTypeStatic,
					wildChild: &node{
						nType:   nodeTypeWild,
						path:    "*",
						handler: mockHandler,
					},
					children: map[string]*node{
						"create": {
							path:     "create",
							nType:    nodeTypeStatic,
							children: map[string]*node{},
							handler:  mockHandler,
						},
					},
				},
				"a": {
					path: "a",
					paramChild: &node{
						path:  ":b",
						nType: nodeTypeParam,
						paramChild: &node{
							path:    ":c",
							nType:   nodeTypeParam,
							handler: mockHandler,
						},
					},
				},
			},
		},
	}}
	msg, ok := r.equal(wantRouter)
	assert.True(t, ok, msg)

	r = newRouter()
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "", mockHandler)
	}, "path can not be empty")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "a", mockHandler)
	}, "path must start with '/'")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/a/", mockHandler)
	}, "path must not end with '/'")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/a///b", mockHandler)
	}, "path must not contain empty path")

	r = newRouter()
	r.addRoute(http.MethodPost, "/", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/", mockHandler)
	}, "have the same route")
	r.addRoute(http.MethodPost, "/a/b", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/a/b", mockHandler)
	}, "have the same route")
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/b", nil)
	}, "handleFunc can not be nil")
	r = newRouter()
	r.addRoute(http.MethodPost, "/order/*", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/order/:order_id", mockHandler)
	}, "both parameters and wildcards can not be registered")
	r.addRoute(http.MethodPost, "/user/:user_id", mockHandler)
	assert.Panicsf(t, func() {
		r.addRoute(http.MethodPost, "/user/*", mockHandler)
	}, "both parameters and wildcards can not be registered")

}

func (r *router) equal(target *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := target.trees[k]
		if !ok {
			return fmt.Sprintf("can not find http method"), false
		}
		if msg, ok := v.equal(dst); !ok {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(target *node) (string, bool) {
	if n.path != target.path {
		return fmt.Sprintf("path not match"), false
	}
	if len(n.children) != len(target.children) {
		return fmt.Sprintf("child node number not match"), false
	}
	if n.nType != target.nType {
		return fmt.Sprintf("node type not match"), false
	}
	if n.paramChild != nil {
		if msg, ok := n.paramChild.equal(target.paramChild); !ok {
			return msg, false
		}
	}
	if n.wildChild != nil {
		if msg, ok := n.wildChild.equal(target.wildChild); !ok {
			return msg, false
		}
	}
	nHandler := reflect.ValueOf(n.handler)
	targetHandler := reflect.ValueOf(target.handler)
	if nHandler != targetHandler {
		return fmt.Sprintf("can not find handler"), false
	}
	for path, v := range n.children {
		dst, ok := target.children[path]
		if !ok {
			return fmt.Sprintf("can not find child node %s", path), false
		}
		if msg, ok := v.equal(dst); !ok {
			return msg, false
		}
	}
	return "", true
}

func Test_router_FindRouter(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodHead, "/"},
		{http.MethodGet, "/a"},
		{http.MethodGet, "/b/c"},
		{http.MethodPost, "/a/b/c"},
		{http.MethodPost, "/order/*"},
		{http.MethodPost, "/*/order/*"},
		{http.MethodPost, "/a/:b/:c"},
	}
	r := newRouter()
	mockHandler := func(c *Context) {

	}
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}
	testCases := []struct {
		name          string
		method        string
		path          string
		wantFound     bool
		wantMatchInfo *matchInfo
	}{
		{
			name:      "found root",
			method:    http.MethodHead,
			path:      "/",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "/",
					nType:   nodeTypeRoot,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "found",
			method:    http.MethodGet,
			path:      "/a",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "a",
					nType:   nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "not found",
			method:    http.MethodGet,
			path:      "/a/b",
			wantFound: false,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "b",
					nType:   nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "no method",
			method:    http.MethodHead,
			path:      "/a/b",
			wantFound: false,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "b",
					nType:   nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "found but no handler",
			method:    http.MethodGet,
			path:      "b",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path: "b",
					children: map[string]*node{
						"c": {
							path:    "c",
							nType:   nodeTypeStatic,
							handler: mockHandler,
						},
					},
				},
			},
		},
		{
			name:      "found wild child",
			method:    http.MethodPost,
			path:      "/order/*",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "*",
					nType:   nodeTypeWild,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "found wild child",
			method:    http.MethodPost,
			path:      "/*/order/*",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "*",
					nType:   nodeTypeWild,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "found param child",
			method:    http.MethodPost,
			path:      "/a/:b/:c",
			wantFound: true,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    ":c",
					nType:   nodeTypeParam,
					handler: mockHandler,
				},
			},
		},
		{
			name:      "do not find child",
			method:    http.MethodDelete,
			path:      "/a",
			wantFound: false,
			wantMatchInfo: &matchInfo{
				n: &node{
					path:    "a",
					nType:   nodeTypeStatic,
					handler: mockHandler,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRouter(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.wantMatchInfo.n.path, n.n.path)
			msg, equal := tc.wantMatchInfo.n.equal(n.n)
			assert.True(t, equal, msg)
		})
	}
}
