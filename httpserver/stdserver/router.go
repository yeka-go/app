package stdserver

import "net/http"

type Middleware func(next http.HandlerFunc) http.HandlerFunc

type Router struct {
	mux *http.ServeMux
	mw  []Middleware
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (router *Router) Group(fn func(router *Router)) {
	fn(&Router{mux: router.mux, mw: router.mw})
}

// Use function registers middleware that will be called for every handler set after it.
func (router *Router) Use(mw ...Middleware) {
	router.mw = append(router.mw, mw...)
}

// HandleFunc registers the handler function (wrapped with registered Middleware) for the given pattern.
// If the given pattern conflicts with one that is already registered, HandleFunc panics.
// The pattern can be "GET /path/{param}" and the param can be retrieved using `r.PathValue("param")`
// where r is *http.Request variable.
func (router *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	for i := len(router.mw) - 1; i >= 0; i-- {
		handler = router.mw[i](handler)
	}
	router.mux.HandleFunc(pattern, handler)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}
