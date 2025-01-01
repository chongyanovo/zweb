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
			path: "/",
			children: map[string]*node{
				"user": {
					path: "user",
					children: map[string]*node{
						"home": {
							path:     "home",
							children: map[string]*node{},
							handler:  mockHandler,
						},
					},
				},
				"order": {
					path: "order",
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
			path: "/",
			wildChild: &node{
				path: "*",
				wildChild: &node{
					path:    "*",
					handler: mockHandler,
				},
				children: map[string]*node{
					"order": {
						path: "order",
						wildChild: &node{
							path:    "*",
							handler: mockHandler,
						},
					},
				},
			},
			children: map[string]*node{
				"order": {
					path: "order",
					wildChild: &node{
						path:    "*",
						handler: mockHandler,
					},
					children: map[string]*node{
						"create": {
							path:     "create",
							children: map[string]*node{},
							handler:  mockHandler,
						},
					},
				},
				"a": {
					path: "a",
					paramChild: &node{
						path: ":b",
						paramChild: &node{
							path:    ":c",
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
		name      string
		method    string
		path      string
		wantFound bool
		wantNode  *node
	}{
		{
			name:      "found",
			method:    http.MethodGet,
			path:      "/a",
			wantFound: true,
			wantNode: &node{
				path:    "a",
				handler: mockHandler,
			},
		},
		{
			name:      "not found",
			method:    http.MethodGet,
			path:      "/a/b",
			wantFound: false,
			wantNode: &node{
				path:    "b",
				handler: mockHandler,
			},
		},
		{
			name:      "no method",
			method:    http.MethodHead,
			path:      "/a/b",
			wantFound: false,
			wantNode: &node{
				path:    "b",
				handler: mockHandler,
			},
		},
		{
			name:      "found but no handler",
			method:    http.MethodGet,
			path:      "b",
			wantFound: true,
			wantNode: &node{
				path: "b",
				children: map[string]*node{
					"c": {
						path:    "c",
						handler: mockHandler,
					},
				},
			},
		},
		{
			name:      "found wild child",
			method:    http.MethodPost,
			path:      "/order/*",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			name:      "found wild child",
			method:    http.MethodPost,
			path:      "/*/order/*",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			name:      "found param child",
			method:    http.MethodPost,
			path:      "/a/:b/:c",
			wantFound: true,
			wantNode: &node{
				path:    ":c",
				handler: mockHandler,
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
			assert.Equal(t, tc.wantNode.path, n.path)
			msg, equal := tc.wantNode.equal(n)
			assert.True(t, equal, msg)
		})
	}
}
