package myhttp

import (
	"slices"
	"strings"
)

var (
	Status200 = []byte("HTTP/1.1 200 OK")
	Status201 = []byte("HTTP/1.1 201 Created")
	Status404 = []byte("HTTP/1.1 404 Not Found")
)

type HandlerFunc func(*Request) *Response

type Router struct {
	tree RouteTrieNode
	//routes         map[string]*Route
	//hasVarSubRoute bool
	//varSubRoute    *Route
}

func NewRouter() *Router {
	tree := RouteTrieNode{
		subRoutes:            make(map[string]*Route),
		hasPathParamSubRoute: false,
		pathParamSubRoute:    nil,
	}
	r := &Router{
		tree,
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
	if route.isVar {
		r.tree.hasPathParamSubRoute = true
		r.tree.pathParamSubRoute = route
	} else {
		r.tree.subRoutes[pathSegments[0]] = route
	}

}

func (r *Router) buildRoute(method string, pathSegments []string, handler HandlerFunc) *Route {
	baseRoute, remainingParts, pathParams := r.findRoute(pathSegments)
	if baseRoute == nil {
		baseRoute = buildSubRoute(method, remainingParts, pathParams, handler)
		return baseRoute
	}

	if remainingParts == nil || len(remainingParts) == 0 {
		setFinalRoute(baseRoute, method, pathParams, handler)
		return baseRoute
	}

	subRoute := buildSubRoute(method, remainingParts, pathParams, handler)
	if subRoute.isVar {
		baseRoute.tree.hasPathParamSubRoute = true
		baseRoute.tree.pathParamSubRoute = subRoute
	} else {
		baseRoute.tree.subRoutes[remainingParts[0]] = subRoute
	}

	return baseRoute
}

// finds the longest existing route matching the path. If the path
// doesn't get matched in full, returns the remaining not-matched parts
func (r *Router) findRoute(pathSegments []string) (route *Route, remainingSegments []string, pathParams []string) {
	pathParams = make([]string, len(pathSegments))

	matchedRoute, exactMatch := r.tree.subRoutes[pathSegments[0]]
	if !exactMatch {
		if isPathParam(pathSegments[0]) && r.tree.hasPathParamSubRoute {
			matchedRoute = r.tree.pathParamSubRoute
			pathParams = append(pathParams, extractPathParam(pathSegments[0]))
		} else {
			return nil, pathSegments, pathParams
		}
	}

	// todo refactor: loop and code above are the same!!
	//  idea: "RouteTrieNode" struct; Router has pointer to Trie root

	for i := 1; i < len(pathSegments); i++ {
		subRoute, exactMatch := matchedRoute.tree.subRoutes[pathSegments[i]]
		if !exactMatch {
			if isPathParam(pathSegments[i]) && matchedRoute.tree.hasPathParamSubRoute {
				subRoute = r.tree.pathParamSubRoute
				pathParams = append(pathParams, extractPathParam(pathSegments[i]))
			} else {
				return matchedRoute, pathSegments[i:], pathParams
			}
		}
		matchedRoute = subRoute
	}

	return matchedRoute, nil, pathParams
}

func buildSubRoute(method string, pathSegments []string, pathParams []string, handler HandlerFunc) *Route {
	route := &Route{}
	if isPathParam(pathSegments[0]) {
		pathParams = append(pathParams, extractPathParam(pathSegments[0]))
	}

	if len(pathSegments) == 1 {
		setFinalRoute(route, method, pathParams, handler)
	} else {
		subRoute := buildSubRoute(method, pathSegments[1:], pathParams, handler)
		route.tree.subRoutes[pathSegments[1]] = subRoute
	}

	return route
}

func setFinalRoute(routeWriter *Route, method string, pathParams []string, handler HandlerFunc) {
	//res := &Route{hasCatchallHandler: true}
	routeWriter.pathParams = pathParams
	routeWriter.tree.isLeafNode = true

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

func isPathParam(pathPart string) bool {
	return strings.HasPrefix(pathPart, "{") && strings.HasSuffix(pathPart, "}")
}

func extractPathParam(pathPart string) string {
	return pathPart[1 : len(pathPart)-1]
}

// func (r *Router) handleRequest(w http.ResponseWriter, req Request) Response { todo
func (r *Router) handleRequest(req *Request) *Response {
	route, context := r.match(req)
	if route == nil {
		return NewResponse().WithStatusLine(Status404)
	}
	req.PathVariables = context

	if !route.tree.isLeafNode {
		panic("Route isn't leaf node!!") // todo avoid this
	}

	var handler HandlerFunc
	if route.hasMethodHandlers {
		handler = route.methodHandlers[req.Method]
	} else {
		handler = route.catchallHandler
	}

	return handler(req)
}

func (r *Router) match(req *Request) (*Route, map[string]string) {
	context := make(map[string]string)

	pathSegments := strings.Split(req.Path, "/")
	route, pathArgs := dfs(r.tree, pathSegments)
	for i, param := range route.pathParams {
		context[param] = pathArgs[i]
	}

	return route, context
}

func dfs(root RouteTrieNode, pathSegments []string) (*Route, []string) {
	// todo there is no need for dfs if we only have 1 node per path param!!!

	if len(pathSegments) == 0 {
		return nil, nil
	}

	if len(pathSegments) == 1 {
		if subRoute, exactMatch := root.subRoutes[pathSegments[0]]; exactMatch {
			return subRoute, nil
		}

		if root.hasPathParamSubRoute {
			subRoute := root.pathParamSubRoute
			return subRoute, []string{pathSegments[0]}
		}

		// neither exact nor param match; subRoute doesn't exist
		return nil, nil
	}

	if subRoute, exactMatch := root.subRoutes[pathSegments[0]]; exactMatch {
		return dfs(subRoute.tree, pathSegments[1:])
	}

	if root.hasPathParamSubRoute {
		matchedSubRoute, vars := dfs(root.pathParamSubRoute.tree, pathSegments[1:])
		vars = slices.Insert(vars, 0, pathSegments[0])
		return matchedSubRoute, vars
	}

	return nil, nil
}

type Route struct {
	tree RouteTrieNode

	isVar bool

	hasCatchallHandler bool
	hasMethodHandlers  bool
	catchallHandler    HandlerFunc
	methodHandlers     map[string]HandlerFunc

	pathParams []string
}

type RouteTrieNode struct {
	subRoutes map[string]*Route

	hasPathParamSubRoute bool
	pathParamSubRoute    *Route

	isLeafNode bool
}
