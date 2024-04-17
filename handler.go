package domainmux

import (
	"net/http"
	"slices"
	"strings"
)

type Handler func(ctx *Context, w http.ResponseWriter, r *http.Request)

type (
	selectFunc func(ctx *Context) bool
	setArgsFunc func(ctx *Context)
)

type handler struct {
	h        Handler
	selectF  selectFunc
	setArgsF setArgsFunc
}

func (h *handler) call(ctx *Context, w http.ResponseWriter, r *http.Request) {
	if !h.selectF(ctx) {
		return
	}

	if h.setArgsF != nil {
		h.setArgsF(ctx)
	}

	h.h(ctx, w, r)
}

func parseQuery(query string) (path []string, selectF selectFunc, setArgsF setArgsFunc) {
	path = strings.Split(query, ".")
	slices.Reverse(path)
	
	switch {
	case query == "*":
		path = nil

		selectF = func(ctx *Context) bool {
			return len(ctx.path) != 0
		}
		
	case path[len(path)-1] == "*":
		path = path[:len(path)-1]

		selectF = func(ctx *Context) bool {
			return len(ctx.path) != 0
		}
		
	default:
		selectF = func(ctx *Context) bool {
			return len(ctx.path) == 0
		}
	}

	return 
}
