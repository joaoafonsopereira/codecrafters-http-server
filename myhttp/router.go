package myhttp

import (
	"strings"
)

var (
	Status200 = []byte("HTTP/1.1 200 OK")
	Status201 = []byte("HTTP/1.1 201 Created")
	Status404 = []byte("HTTP/1.1 404 Not Found")
)

type HandlerFunc func(*Request) *Response

type Router struct {
	routes map[string]*Route
	// echo/{str}
	// / ""
}

func NewRouter() *Router {
	r := &Router{
		routes: make(map[string]*Route),
	}
	return r
}

func (r *Router) RegisterHandler(pattern string, handler HandlerFunc) {
	var method string
	var path string
	patternParts := strings.Split(pattern, " ")
	if len(patternParts) == 2 {
		method = patternParts[0]
		path = patternParts[1]
	} else {
		method = ""
		path = patternParts[0]
	}

	pathSegments := strings.Split(path, "/")
	if len(pathSegments) > 1 {
		pathSegments = pathSegments[1:] // parts[0] is always "" and doesn't matter in these cases
	}

	route := r.buildRoute(method, pathSegments, handler)
	r.routes[pathSegments[0]] = route

}

func (r *Router) buildRoute(method string, pathSegments []string, handler HandlerFunc) *Route {
	baseRoute, remainingParts := r.findRoute(pathSegments)
	if baseRoute == nil {
		baseRoute = buildSubRoute(method, remainingParts, handler)
		return baseRoute
	}

	if remainingParts == nil || len(remainingParts) == 0 {
		setFinalRoute(baseRoute, method, handler)
		return baseRoute
	}

	subRoute := buildSubRoute(method, remainingParts, handler)
	baseRoute.subRoutes[remainingParts[0]] = subRoute
	return baseRoute
}

// finds the longest existing route matching the path. If the path
// doesn't get matched in full, returns the remaining not-matched parts
func (r *Router) findRoute(pathSegments []string) (*Route, []string) {
	matchedRoute, ok := r.routes[pathSegments[0]]
	if !ok {
		return nil, pathSegments
	}

	// how to handle variables?
	// for now, assume that a path variable always has the same name

	for i := 1; i < len(pathSegments); i++ {
		subRoute, exists := matchedRoute.subRoutes[pathSegments[i]]
		if !exists {
			// todo check if it's a variable, and if so allow a different variable name???

			return matchedRoute, pathSegments[i:]
		}
		matchedRoute = subRoute
	}

	return matchedRoute, nil
}

func buildSubRoute(method string, pathSegments []string, handler HandlerFunc) *Route {
	route := &Route{}
	if isPathVar(pathSegments[0]) {
		route.isVar = true
		route.varName = extractPathVarName(pathSegments[0])
	}

	if len(pathSegments) == 1 {
		setFinalRoute(route, method, handler)
	} else {
		subRoute := buildSubRoute(method, pathSegments[1:], handler)
		route.subRoutes[pathSegments[1]] = subRoute
	}

	return route
}

func setFinalRoute(routeWriter *Route, method string, handler HandlerFunc) {
	//res := &Route{hasCatchallHandler: true}
	if method == "" {
		routeWriter.hasCatchallHandler = true
		routeWriter.catchallHandler = handler
	} else {
		//routeWriter.catchallHandler = nil todo do we want to remove the catchall catchallHandler?
		routeWriter.hasMethodHandlers = true
		routeWriter.methodHandlers[method] = handler
	}
	//return res
}

func isPathVar(pathPart string) bool {
	return strings.HasPrefix(pathPart, "{") && strings.HasSuffix(pathPart, "}")
}

func extractPathVarName(pathPart string) string {
	return pathPart[1 : len(pathPart)-1]
}

// func (r *Router) handleRequest(w http.ResponseWriter, req Request) Response { todo
func (r *Router) handleRequest(req *Request) *Response {
	if handle, ok := r.routes[req.methodAndPath]; ok {
		return handle(req)
	}
	if handle, ok := r.routes[req.Path]; ok {
		return handle(req)
	}

	return NewResponse().WithStatusLine(Status404)
}
