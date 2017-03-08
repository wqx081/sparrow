package sparrow

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"sync"
)

type Params map[string]string

//
type Router interface {
	Routes

	// Group 增加一个组, 该组中所有关联的 routes 同样被添加
	Group(string, func(Router), ...Handler)
	// GET 添加一个 HTTP GET 请求的 route
	GET(string, ...Handler) Route

	NotFound(...Handler)

	// HTTP 处理入口
	Handle(http.ResponseWriter, *http.Request, Context)
}

type router struct {
	routes     []*route
	notFounds  []Handler
	groups     []group
	routesLock sync.RWMutex
}

type group struct {
	pattern  string
	handlers []Handler
}

func NewRouter() Router {
	return &router{notFounds: []Handler{http.NotFound}, groups: make([]group, 0)}
}

func (r *router) Group(pattern string, fn func(Router), h ...Handler) {
	r.groups = append(r.groups, group{pattern, h})
	fn(r)
	r.groups = r.groups[:len(r.groups)-1]
}

func (r *router) GET(pattern string, h ...Handler) Route {
	return r.addRoute("GET", pattern, h)
}

func (r *router) Any(pattern string, h ...Handler) Route {
	return r.addRoute("*", pattern, h)
}

func (r *router) AddRoute(method, pattern string, h ...Handler) Route {
	return r.addRoute(method, pattern, h)
}

//TODO(wqx)

func (r *router) addRoute(method string, pattern string, handlers []Handler) *route {
	if len(r.groups) > 0 {
	}

	route := newRoute(method, pattern, handlers)
	route.Validate()
	r.appendRoute(route)
	return route
}

///

// 表示在 routing 层的 Route
type Route interface {
	URLWith([]string) string
	Name(string)
	GetName() string
	Pattern() string
	Method() string
}

type route struct {
	method   string
	regex    *regexp.Regexp
	handlers []Handler
	pattern  string
	name     string
}

// Helper

type Routes interface {
	URLFor(name string, params ...interface{}) string
	MethodsFor(path string) []string
	All() []Route
}

//
type routeContext struct {
	Context
	index    int
	handlers []Handler
}
