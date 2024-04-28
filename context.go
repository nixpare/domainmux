package domainmux

import (
	"net/http"
	"slices"
	"strings"
)

type Context struct {
	dm             *DomainMux
	host           string
	path           []string
	args           map[string]string
	node           *node
	handlers       []*nodeHandler
	calledHandlers map[*nodeHandler]struct{}
	serveCalled    bool
}

func newContext(dm *DomainMux, host string) *Context {
	ctx := &Context{
		dm: dm,
		args: make(map[string]string),
		calledHandlers: make(map[*nodeHandler]struct{}),
	}

	ctx.setup(host)
	return ctx
}

func (ctx *Context) setup(host string) {
	ctx.host = host

	ctx.path = append(strings.Split(host, "."), "")
	slices.Reverse(ctx.path)

	ctx.node = ctx.dm.root
}

func (ctx *Context) Path() []string {
	cp := make([]string, len(ctx.path))
	copy(cp, ctx.path)
	return cp
}

func (ctx *Context) DomainMux() *DomainMux {
	return ctx.dm
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

func (ctx *Context) Value(key string) string {
	return ctx.args[key]
}

func (ctx *Context) HasValue(key string) bool {
	_, ok := ctx.args[key]
	return ok
}

func (ctx *Context) SetValue(key, value string) {
	ctx.args[key] = value
}

func (ctx *Context) IsServeCalled() bool {
	return ctx.serveCalled
}

func (ctx *Context) SetServeCalled() {
	ctx.serveCalled = true
}

func (ctx *Context) Redirect(host string, rerun bool) {
	ctx.setup(host)
	if rerun {
		clear(ctx.calledHandlers)
	}
}

func (ctx *Context) next() *nodeHandler {
	if ctx.serveCalled {
		return nil
	}

	if len(ctx.handlers) > 0 {
		h := ctx.handlers[0]
		ctx.handlers = ctx.handlers[1:]

		if _, ok := ctx.calledHandlers[h]; !ok {
			ctx.calledHandlers[h] = struct{}{}
		} else {
			h = ctx.next()
		}

		return h
	}

	if ctx.node == nil {
		return nil
	}

	ctx.handlers = ctx.node.handlers

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
