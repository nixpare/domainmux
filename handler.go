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
	selectF  []selectFunc
	setArgsF []setArgsFunc
}

func (h *handler) call(ctx *Context, w http.ResponseWriter, r *http.Request) {
	for _, selectF := range h.selectF {
		if !selectF(ctx) {
			return
		}
	}

	for _, setArgsF := range h.setArgsF {
		setArgsF(ctx)
	}

	h.h(ctx, w, r)
}

func parseQuery(query string) (path []string, selectF []selectFunc, setArgsF []setArgsFunc) {
	if query == "*" {
		selectF = append(selectF, func(ctx *Context) bool {
			return len(ctx.path) != 0
		})

		path = nil
		return
	}

	if strings.HasPrefix(query, "...") {
		path = strings.Split(strings.TrimLeft(query, "."), ".")
		path[0] = "..." + path[0]
	} else {
		path = strings.Split(query, ".")
	}
	slices.Reverse(path)

	endPath := len(path)

	for i, p := range path {
		switch {			
		case p == "*" && i == len(path)-1:
			selectF = append(selectF, func(ctx *Context) bool {
				return len(ctx.path) != 0
			})
	
			endPath --
	
		case p == "*?" && i == len(path)-1:
			selectF = append(selectF, func(ctx *Context) bool {
				return true
			})
	
			endPath --
		
		case strings.HasPrefix(p, ":"):
			key := strings.TrimLeft(p, ":")
			index := i
	
			if strings.HasSuffix(p, "?") {
				key = strings.TrimRight(key, "?")
	
				selectF = append(selectF, func(ctx *Context) bool {
					return len(ctx.path) == 1 || len(ctx.path) == 0
				})
			} else {
				selectF = append(selectF, func(ctx *Context) bool {
					return len(ctx.path) >= index
				})
			}
	
			if key != "_" {
				setArgsF = append(setArgsF, func(ctx *Context) {
					if len(ctx.path) >= index {
						println(index, ctx.path[index])
						ctx.args[key] = ctx.path[index]
					} else {
						ctx.args[key] = ""
					}
				})
			}
	
			endPath --
	
		case strings.HasPrefix(p, "...") && i == len(path)-1:
			key := strings.TrimLeft(p, ".")
	
			if strings.HasSuffix(p, "?") {
				key = strings.TrimRight(key, "?")
	
				selectF = append(selectF, func(ctx *Context) bool {
					return true
				})
			} else {
				selectF = append(selectF, func(ctx *Context) bool {
					return len(ctx.path) != 0
				})
			}
	
			if key != "_" {
				setArgsF = append(setArgsF, func(ctx *Context) {
					pathCopy := make([]string, len(ctx.path))
					copy(pathCopy, ctx.path)
					slices.Reverse(pathCopy)
					ctx.args[key] = strings.Join(pathCopy, ".")
				})
			}
	
			endPath --

		case i == len(path)-1:
			selectF = append(selectF, func(ctx *Context) bool {
				return len(ctx.path) == 0
			})
		}
	}

	path = path[:endPath]
	return 
}
