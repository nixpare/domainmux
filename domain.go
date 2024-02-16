package domainmux

import (
	"net/http"
	"slices"
	"strings"
)

type DomainMux nodeMap

func NewDomainMux() *DomainMux {
	return (*DomainMux)(&nodeMap{
		m: make(map[string]*Node),
	})
}

func (dm *DomainMux) Serve(host string, middlewares ...Handler) {
	var path []string

	if strings.HasPrefix(host, "...") {
		host = strings.TrimLeft(host, ".")
		path = strings.Split(host, ".")
		path[0] = "..." + path[0]
	} else {
		path = strings.Split(host, ".")
	}

	slices.Reverse(path)

	switch {
	case strings.HasPrefix(path[len(path)-1], "?"):
		opt := strings.TrimLeft(path[len(path)-1], "?")
		path[len(path)-1] = ":" + opt

		shortMWs := make([]Handler, 0, len(middlewares))
		resetOpt := strings.TrimLeft(opt, ".:")

		if resetOpt != "_" {
			for _, h := range middlewares {
				shortMWs = append(shortMWs, func(ctx *Context, w http.ResponseWriter, r *http.Request) {
					ctx.args[resetOpt] = ""
					h(ctx, w, r)
				})
			}
		} else {
			shortMWs = middlewares
		}

		(*nodeMap)(dm).createNode(path[:len(path)-1], 0, shortMWs)
		(*nodeMap)(dm).createNode(path, 0, middlewares)
	case strings.HasPrefix(path[len(path)-1], "...?"):
		opt := strings.Replace(path[len(path)-1], "?", "", 1)
		path[len(path)-1] = opt

		shortMWs := make([]Handler, 0, len(middlewares))
		resetOpt := strings.TrimLeft(opt, ".:")

		for _, h := range middlewares {
			shortMWs = append(shortMWs, func(ctx *Context, w http.ResponseWriter, r *http.Request) {
				ctx.args[resetOpt] = ""
				h(ctx, w, r)
			})
		}

		(*nodeMap)(dm).createNode(path[:len(path)-1], 0, shortMWs)
		(*nodeMap)(dm).createNode(path, 0, middlewares)
	default:
		(*nodeMap)(dm).createNode(path, 0, middlewares)
	}
}

func (dm *DomainMux) Execute(host string, w http.ResponseWriter, r *http.Request) {
	ctx := newContext(dm, host)
	ctx.Next(w, r)
}

func (dm *DomainMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dm.Execute(SplitAddrPort(r.Host), w, r)
}

func (dm *DomainMux) String() string {
	sb := strings.Builder{}
	sb.WriteString("Domains:\n")

	for _, key := range dm.keys {
		sb.WriteString(indentString(dm.m[key].printNode(), 4))
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}
