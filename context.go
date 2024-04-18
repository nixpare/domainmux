package domainmux

import (
	"net/http"
	"slices"
	"strings"
)

type Context struct {
	host       string
	path       []string
	args       map[string]string
	node       *node
	handlers   []*handler
}

func newContext(dm *DomainMux, host string) *Context {
	path := append(strings.Split(host, "."), "")
	slices.Reverse(path)

	ctx := &Context{
		host: host,
		path: path,
		args: make(map[string]string),
		node: dm.root,
	}
	return ctx
}

func (ctx *Context) Host() string {
	return ctx.host
}

func (ctx *Context) Param(key string) string {
	return ctx.args[key]
}

func (ctx *Context) HasParam(key string) bool {
	_, ok := ctx.args[key]
	return ok
}

func (ctx *Context) Value(key string) (string, bool) {
	value, ok := ctx.args[key]
	return value, ok
}

func (ctx *Context) next() *handler {
	if len(ctx.handlers) > 0 {
		h := ctx.handlers[0]
		ctx.handlers = ctx.handlers[1:]
		return h
	}

	if ctx.node == nil {
		return nil
	}

	if len(ctx.node.mws) > 0 {
		ctx.handlers = ctx.node.mws
	}

	if len(ctx.path) > 1 {
        ctx.path = ctx.path[1:]
        ctx.node = ctx.node.childs[ctx.path[0]]
	} else {
        ctx.path = nil
        ctx.node = nil
    }
	
	return ctx.next()
}

func (ctx *Context) Next(w http.ResponseWriter, r *http.Request) {
	for h := ctx.next(); h != nil; h = ctx.next() {
		h.call(ctx, w, r)
	}
}
