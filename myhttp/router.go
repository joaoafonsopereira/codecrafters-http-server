package myhttp

var (
	Status200 = []byte("HTTP/1.1 200 OK")
	Status201 = []byte("HTTP/1.1 201 Created")
	Status404 = []byte("HTTP/1.1 404 Not Found")
)

type HandlerFunc func(*Request) *Response

type Router struct {
	routes map[string]HandlerFunc
}

func NewRouter() *Router {
	r := &Router{
		routes: make(map[string]HandlerFunc),
	}
	return r
}

func (r *Router) RegisterHandler(pattern string, handler HandlerFunc) {
	r.routes[pattern] = handler
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
