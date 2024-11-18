package myhttp

import (
	"slices"
	"strings"
)

type HandlerFunc func(ResponseWriter, *Request)

type Router struct {
	tree *RouteTrieNode
}

func NewRouter() *Router {
	return &Router{
		NewRouteTrieNode(),
	}
}

type RouteTrieNode struct {
	subRoutes map[string]*Route

	hasPathParamSubRoute bool
	pathParamSubRoute    *Route
	isLeafNode           bool
}

func NewRouteTrieNode() *RouteTrieNode {
	return &RouteTrieNode{
		subRoutes:            make(map[string]*Route),
		hasPathParamSubRoute: false,
		pathParamSubRoute:    nil,
	}
}

type Route struct {
	tree *RouteTrieNode

	isVar      bool
	pathParams []string

	hasCatchallHandler bool
	hasMethodHandlers  bool
	catchallHandler    HandlerFunc
	methodHandlers     map[string]HandlerFunc
}

func NewRoute() *Route {
	return &Route{
		tree:           NewRouteTrieNode(),
		methodHandlers: make(map[string]HandlerFunc),
	}
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
	addSubRoute(r.tree, route, pathSegments[0])
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
	addSubRoute(baseRoute.tree, subRoute, remainingParts[0])

	return baseRoute
}

func addSubRoute(root *RouteTrieNode, subRoute *Route, subRouteFirstSegment string) {
	if subRoute.isVar {
		root.hasPathParamSubRoute = true
		root.pathParamSubRoute = subRoute
	} else {
		root.subRoutes[subRouteFirstSegment] = subRoute
	}
}

// finds the longest existing route matching the path. If the path
// doesn't get matched in full, returns the remaining not-matched parts
func (r *Router) findRoute(pathSegments []string) (route *Route, remainingSegments []string, pathParams []string) {
	pathParams = make([]string, 0, len(pathSegments))

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
				subRoute = matchedRoute.tree.pathParamSubRoute
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
	route := NewRoute()
	if isPathParam(pathSegments[0]) {
		pathParams = append(pathParams, extractPathParam(pathSegments[0]))
		route.isVar = true
	}

	if len(pathSegments) == 1 {
		setFinalRoute(route, method, pathParams, handler)
	} else {
		subRoute := buildSubRoute(method, pathSegments[1:], pathParams, handler)
		addSubRoute(route.tree, subRoute, pathSegments[1])
	}

	return route
}

func setFinalRoute(routeWriter *Route, method string, pathParams []string, handler HandlerFunc) {
	routeWriter.pathParams = pathParams
	routeWriter.tree.isLeafNode = true

	if method == "" {
		routeWriter.hasCatchallHandler = true
		routeWriter.catchallHandler = handler
	} else {
		routeWriter.hasMethodHandlers = true
		routeWriter.methodHandlers[method] = handler
	}
}

func isPathParam(pathPart string) bool {
	return strings.HasPrefix(pathPart, "{") && strings.HasSuffix(pathPart, "}")
}

func extractPathParam(pathPart string) string {
	return pathPart[1 : len(pathPart)-1]
}

func (r *Router) match(req *Request) (HandlerFunc, map[string]string) {
	pathSegments := strings.Split(req.Path, "/")[1:] // first segment is always "" and doesn't matter
	route, pathArgs := matchPathRec(r.tree, pathSegments)
	if route == nil {
		return nil, nil
	}
	if !route.tree.isLeafNode {
		panic("Route isn't leaf node!!") // todo avoid this
	}
	if !route.hasCatchallHandler && !route.hasMethodHandlers {
		panic("Route has no corresponding handler!!")
	}

	var handler HandlerFunc
	if route.hasMethodHandlers {
		handler = route.methodHandlers[req.Method]
	} else {
		handler = route.catchallHandler
	}

	if len(pathArgs) != len(route.pathParams) {
		panic("Discovered path arguments are different than declared path parameters!")
	}
	context := make(map[string]string)
	for i, param := range route.pathParams {
		context[param] = pathArgs[i]
	}

	return handler, context
}

func matchPathRec(root *RouteTrieNode, pathSegments []string) (*Route, []string) {
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
		return matchPathRec(subRoute.tree, pathSegments[1:])
	}

	if root.hasPathParamSubRoute {
		matchedSubRoute, vars := matchPathRec(root.pathParamSubRoute.tree, pathSegments[1:])
		vars = slices.Insert(vars, 0, pathSegments[0])
		return matchedSubRoute, vars
	}

	return nil, nil
}
