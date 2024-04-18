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
	if strings.HasPrefix(query, "...") {
		path = strings.Split(strings.TrimLeft(query, "."), ".")
		path[0] = "..." + path[0]
	} else {
		path = strings.Split(query, ".")
	}
	slices.Reverse(path)
	
	switch {
	case query == "*":
		selectF = func(ctx *Context) bool {
			return len(ctx.path) != 0
		}

		path = nil
		
	case path[len(path)-1] == "*":
		selectF = func(ctx *Context) bool {
			return len(ctx.path) != 0
		}

		path = path[:len(path)-1]
	
	case strings.HasPrefix(path[len(path)-1], ":"):
		key := strings.TrimLeft(path[len(path)-1], ":")

		if strings.HasSuffix(path[len(path)-1], "?") {
			key = strings.TrimRight(key, "?")

			selectF = func(ctx *Context) bool {
				return len(ctx.path) == 1 || len(ctx.path) == 0
			}
		} else {
			selectF = func(ctx *Context) bool {
				return len(ctx.path) == 1
			}
		}

		if key != "_" {
			setArgsF = func(ctx *Context) {
				if len(ctx.path) == 1 {
					ctx.args[key] = ctx.path[0]
				} else {
					ctx.args[key] = ""
				}
			}
		}

		path = path[:len(path)-1]

	case strings.HasPrefix(path[len(path)-1], "..."):
		key := strings.TrimLeft(path[len(path)-1], ".")

		if strings.HasSuffix(path[len(path)-1], "?") {
			key = strings.TrimRight(key, "?")

			selectF = func(ctx *Context) bool {
				return true
			}
		} else {
			selectF = func(ctx *Context) bool {
				return len(ctx.path) != 0
			}
		}

		if key != "_" {
			setArgsF = func(ctx *Context) {
				pathCopy := make([]string, len(ctx.path))
				copy(pathCopy, ctx.path)
				slices.Reverse(pathCopy)
				ctx.args[key] = strings.Join(pathCopy, ".")
			}
		}

		path = path[:len(path)-1]
		
	default:
		selectF = func(ctx *Context) bool {
			return len(ctx.path) == 0
		}
	}

	return 
}
