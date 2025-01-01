package test

import (
	"fmt"
	"github.com/chongyanovo/zweb"
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
	}
	mockHandler := func(c *main.Context) {

	}
	r := main.newRouter()
	for _, test := range testRoutes {
		r.addRoute(test.method, test.path, mockHandler)
	}
	wantRouter := &main.router{trees: map[string]*main.node{
		http.MethodGet: {
			path: "/",
			children: map[string]*main.node{
				"user": {
					path: "user",
					children: map[string]*main.node{
						"home": {
							path:     "home",
							children: map[string]*main.node{},
							handler:  mockHandler,
						},
					},
				},
				"order": {
					path: "order",
					children: map[string]*main.node{
						"detail": {
							path:     "detail",
							children: map[string]*main.node{},
							handler:  mockHandler,
						},
					},
				},
			},
			handler: mockHandler,
		},
		http.MethodPost: {
			path: "/",
			children: map[string]*main.node{
				"order": {
					path: "order",
					children: map[string]*main.node{
						"create": {
							path:     "create",
							children: map[string]*main.node{},
							handler:  mockHandler,
						},
					},
				},
			},
		},
	}}
	msg, ok := r.equal(wantRouter)
	assert.True(t, ok, msg)

	r = main.newRouter()
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

	r = main.newRouter()
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
}

func (r *main.router) equal(target *main.router) (string, bool) {
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

func (n *main.node) equal(target *main.node) (string, bool) {
	if n.path != target.path {
		return fmt.Sprintf("path不匹配"), false
	}
	if len(n.children) != len(target.children) {
		return fmt.Sprintf("child node number not match"), false
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
	}
	r := main.newRouter()
	mockHandler := func(c *main.Context) {

	}
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}
	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		wantNode  *main.node
	}{
		{
			name:      "found",
			method:    http.MethodGet,
			path:      "/a",
			wantFound: true,
			wantNode: &main.node{
				path:    "a",
				handler: mockHandler,
			},
		},
		{
			name:      "not found",
			method:    http.MethodGet,
			path:      "/a/b",
			wantFound: false,
			wantNode: &main.node{
				path:    "b",
				handler: mockHandler,
			},
		},
		{
			name:      "no method",
			method:    http.MethodHead,
			path:      "/a/b",
			wantFound: false,
			wantNode: &main.node{
				path:    "b",
				handler: mockHandler,
			},
		},
		{
			name:      "found but no handler",
			method:    http.MethodGet,
			path:      "b",
			wantFound: true,
			wantNode: &main.node{
				path: "b",
				children: map[string]*main.node{
					"c": {
						path:     "c",
						children: nil,
						handler:  mockHandler,
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.FindRouter(tc.method, tc.path)
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
