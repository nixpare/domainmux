package domainmux

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type Handler func(ctx *Context, w http.ResponseWriter, r *http.Request)

type (
	selectFunc  func(ctx *Context) bool
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

func parseQuery(query string) (path []string, selectF []selectFunc, setArgsF []setArgsFunc, err error) {
	if query == "" {
		err = fmt.Errorf("empty query")
		return
	}

	if query == "*" {
		selectF = append(selectF, func(ctx *Context) bool {
			return len(ctx.path) != 0
		})

		path = nil
		return
	}

	switch strings.Count(query, "*") {
	case 0:
	case 1:
		if !strings.HasPrefix(query, "*") {
			err = fmt.Errorf("catch-all query must start with \"*\"")
			return
		}

	default:
		err = fmt.Errorf("catch-all query must have at most one \"*\"")
		return
	}

	switch strings.Count(query, "...") {
	case 0:
		path = strings.Split(query, ".")
	case 1:
		if !strings.HasPrefix(query, "...") {
			err = fmt.Errorf("variadic query must start with \"...\"")
			return
		}

		path = strings.Split(strings.TrimLeft(query, "."), ".")
		path[0] = "..." + path[0]
	default:
		err = fmt.Errorf("variadic query must have at most one \"...\"")
		return
	}

	slices.Reverse(path)
	endPath := len(path)

	var paramFound bool

	for i, p := range path {
		var e expr

		var isOptional bool
		if strings.HasSuffix(p, "?") {
			isOptional = true
			p = strings.TrimSuffix(p, "?")
		}

		switch {
		case strings.HasPrefix(p, "*"):
			paramFound = true
			e = &starExpr{
				index: len(path) - endPath,
			}

		case strings.HasPrefix(p, "$"):
			paramFound = true
			e = &paramExpr{
				key:   strings.TrimLeft(p, "$"),
				index: len(path) - endPath,
			}

		case strings.HasPrefix(p, "..."):
			paramFound = true
			e = &variadicExpr{
				key:   strings.TrimLeft(p, "."),
				index: len(path) - endPath,
			}

		default:
			e = &literalExpr{
				name:         p,
				index:        len(path) - endPath,
				isAfterParam: paramFound,
				isLast:       i == len(path)-1,
			}
		}

		if isOptional {
			selectF = append(selectF, e.selectFopt)
		} else {
			selectF = append(selectF, e.selectF)
		}

		setArgsF = append(setArgsF, e.setArgsF)

		if e.decrementPath() {
			endPath--
		}
	}

	path = path[:endPath]
	return
}
