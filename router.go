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

	// GET 添加一个 HTTP GET 请求的 route
	GET(string, ...Handler) Route

	NotFound(...Handler)

	// HTTP 处理入口
	Handle(http.ResponseWriter, *http.Request, Context)
}

type router struct {
	routes     []*route
	notFounds  []Handler
	routesLock sync.RWMutex
}

func NewRouter() Router {
	return &router{notFounds: []Handler{http.NotFound}}
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

func (r *router) Handle(w http.ResponseWriter, req *http.Request, context Context) {
	bestMatch := NoMatch
	var bestVals map[string]string
	var bestRoute *route
	for _, route := range r.getRoutes() {
		match, vals := route.Match(req.Method, req.URL.Path)
		if match.BetterThan(bestMatch) {
			bestMatch = match
			bestVals = vals
			bestRoute = route
			if match == ExactMatch {
				break
			}
		}
	}
	if bestMatch != NoMatch {
		params := Params(bestVals)
		context.Map(params)
		bestRoute.Handle(context, w)
		return
	}

	// No routes exist, 404
	c := &routeContext{context, 0, r.notFounds}
	context.MapTo(c, (*Context)(nil))
	c.run()
}

func (r *router) NotFound(handler ...Handler) {
	r.notFounds = handler
}

func (r *router) addRoute(method string, pattern string, handlers []Handler) *route {
	route := newRoute(method, pattern, handlers)
	route.Validate()
	r.appendRoute(route)
	return route
}

func (r *router) appendRoute(nr *route) {
	r.routesLock.Lock()
	defer r.routesLock.Unlock()
	r.routes = append(r.routes, nr)
}

func (r *router) getRoutes() []*route {
	r.routesLock.RLock()
	defer r.routesLock.RUnlock()
	return r.routes[:]
}

func (r *router) findRoute(name string) *route {
	for _, route := range r.getRoutes() {
		if route.name == name {
			return route
		}
	}

	return nil
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

var routeReg1 = regexp.MustCompile(`:[^/#?()\.\\]+`)
var routeReg2 = regexp.MustCompile(`\*\*`)

func newRoute(method string, pattern string, handlers []Handler) *route {
	route := route{method, nil, handlers, pattern, ""}
	pattern = routeReg1.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	var index int
	pattern = routeReg2.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`
	route.regex = regexp.MustCompile(pattern)
	return &route
}

type RouteMatch int

const (
	NoMatch RouteMatch = iota
	StarMatch
	OverloadMatch
	ExactMatch
)

func (r RouteMatch) BetterThan(o RouteMatch) bool {
	return r > 0
}

func (r route) MatchMethod(method string) RouteMatch {
	switch {
	case method == r.method:
		return ExactMatch
	case method == "HEAD" && r.method == "GET":
		return OverloadMatch
	case method == "*":
		return StarMatch
	default:
		return NoMatch
	}
}

func (r route) Match(method string, path string) (RouteMatch, map[string]string) {
	match := r.MatchMethod(method)
	if match == NoMatch {
		return match, nil
	}

	matches := r.regex.FindStringSubmatch(path)
	if len(matches) > 0 && matches[0] == path {
		params := make(map[string]string)
		for i, name := range r.regex.SubexpNames() {
			if len(name) > 0 {
				params[name] = matches[i]
			}
		}
		return match, params
	}
	return NoMatch, nil
}

func (r *route) Validate() {
	for _, handler := range r.handlers {
		validateHandler(handler)
	}
}

func (r *route) Handle(c Context, w http.ResponseWriter) {
	context := &routeContext{c, 0, r.handlers}
	c.MapTo(context, (*Context)(nil))
	c.MapTo(r, (*Route)(nil))
	context.run()
}

var urlReg = regexp.MustCompile(`:[^/#?()\.\\]+|\(\?P<[a-zA-Z0-9]+>.*\)`)

func (r *route) URLWith(args []string) string {
	if len(args) > 0 {
		argCount := len(args)
		i := 0
		url := urlReg.ReplaceAllStringFunc(r.pattern, func(m string) string {
			var val interface{}
			if i < argCount {
				val = args[i]
			} else {
				val = m
			}
			i += 1
			return fmt.Sprintf(`%v`, val)
		})

		return url
	}
	return r.pattern
}

func (r *route) Name(name string) {
	r.name = name
}

func (r *route) GetName() string {
	return r.name
}

func (r *route) Pattern() string {
	return r.pattern
}

func (r *route) Method() string {
	return r.method
}

// Helper

type Routes interface {
	URLFor(name string, params ...interface{}) string
	MethodsFor(path string) []string
	All() []Route
}

func (r *router) URLFor(name string, params ...interface{}) string {
	route := r.findRoute(name)

	if route == nil {
		panic("route not found")
	}

	var args []string
	for _, param := range params {
		switch v := param.(type) {
		case int:
			args = append(args, strconv.FormatInt(int64(v), 10))
		case string:
			args = append(args, v)
		default:
			if v != nil {
				panic("Arguments passed to URLFor must be integers or strings")
			}
		}
	}

	return route.URLWith(args)
}

func (r *router) All() []Route {
	routes := r.getRoutes()
	var ri = make([]Route, len(routes))

	for i, route := range routes {
		ri[i] = Route(route)
	}

	return ri
}

func hasMethod(methods []string, method string) bool {
	for _, v := range methods {
		if v == method {
			return true
		}
	}
	return false
}

func (r *router) MethodsFor(path string) []string {
	methods := []string{}
	for _, route := range r.getRoutes() {
		matches := route.regex.FindStringSubmatch(path)
		if len(matches) > 0 && matches[0] == path && !hasMethod(methods, route.method) {
			methods = append(methods, route.method)
		}
	}
	return methods
}

//
type routeContext struct {
	Context
	index    int
	handlers []Handler
}

func (r *routeContext) Next() {
	r.index += 1
	r.run()
}

func (r *routeContext) run() {
	for r.index < len(r.handlers) {
		handler := r.handlers[r.index]
		vals, err := r.Invoke(handler)
		if err != nil {
			panic(err)
		}
		r.index += 1

		if len(vals) > 0 {
			ev := r.Get(reflect.TypeOf(ReturnHandler(nil)))
			handleReturn := ev.Interface().(ReturnHandler)
			handleReturn(r, vals)
		}

		if r.Written() {
			return
		}
	}
}
