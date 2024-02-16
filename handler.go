package domainmux

import (
	"net/http"
	"slices"
	"strings"
)

type Handler func(ctx *Context, w http.ResponseWriter, r *http.Request)

type Context struct {
	w    http.ResponseWriter
	r    *http.Request
	host string
	dm   *DomainMux
	args map[string]string
	hnds []Handler
}

func (ctx *Context) W() http.ResponseWriter {
	return ctx.w
}

func (ctx *Context) R() *http.Request {
	return ctx.r
}

func (ctx *Context) Host() string {
	return ctx.host
}

func (ctx *Context) Value(key string) string {
	return ctx.args[key]
}

func (ctx *Context) next() Handler {
	if len(ctx.hnds) == 0 {
		return nil
	}

	h := ctx.hnds[0]
	ctx.hnds = ctx.hnds[1:]
	return h
}

func (ctx *Context) Next(w http.ResponseWriter, r *http.Request) {
	for h := ctx.next(); h != nil; h = ctx.next() {
		h(ctx, w, r)
	}
}

func (ctx *Context) ChangeHost(host string) {
	host = strings.Trim(host, ".")
	ctx.hnds = ctx.hnds[:0]
	ctx.fetchHandlers(host)
}

func newContext(dm *DomainMux, host string) *Context {
	ctx := &Context{ args: make(map[string]string), dm: dm }
	ctx.fetchHandlers(host)
	return ctx
}

func (ctx *Context) fetchHandlers(host string) {
	ctx.host = host
	
	path := strings.Split(host, ".")
	slices.Reverse(path)
	
	var fetch []*fetchRes
	for _, key := range ctx.dm.keys {
		value := ctx.dm.m[key]
		fetch = append(fetch, value.fetchHandlers(path, 0))
	}

	nextFetch := fetch

	for len(nextFetch) != 0 {
		nfl := len(nextFetch)
		for _, res := range nextFetch {
			ctx.hnds = append(ctx.hnds, res.mws...)
		}
		for i := 0; i < nfl; i++ {
			nextFetch = append(nextFetch, nextFetch[i].childs...)
		}
		nextFetch = nextFetch[nfl:]
	}
}
