package domainmux

import (
	"fmt"
	"net/http"
	"strings"
)

type DomainMux struct {
	root *node
}

func NewDomainMux() *DomainMux {
	return &DomainMux{
		root: &node{
			path: "*",
			childs: make(map[string]*node),
		},
	}
}

func (dm *DomainMux) Serve(pattern string, handler Handler) {
	path, selectF, setArgsF, err := parseQuery(pattern)
	if err != nil {
		panic(fmt.Errorf("invalid query: %w", err))
	}

	dm.root.createNode(path, &nodeHandler{
		handler: handler,
		isServe: true,
		selectF: selectF,
		setArgsF: setArgsF,
	})
}

func (dm *DomainMux) Middleware(pattern string, middlewares ...Handler) {
	path, selectF, setArgsF, err := parseQuery(pattern)
	if err != nil {
		panic(fmt.Errorf("invalid query: %w", err))
	}

	handlers := make([]*nodeHandler, 0, len(middlewares))
	for _, mw := range middlewares {
		handlers = append(handlers, &nodeHandler{
			handler: mw,
			isServe: false,
			selectF: selectF,
			setArgsF: setArgsF,
		})
	}

	dm.root.createNode(path, handlers...)
}

func (dm *DomainMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(dm, SplitAddrPort(r.Host))
	ctx.Next(w, r)

	if !ctx.IsServeCalled() {
		ctx.SetServeCalled()
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Host %s not served by this server", ctx.Host())))
	}
}

func (dm *DomainMux) String() string {
	sb := strings.Builder{}
	sb.WriteString("Domains:\n")

	sb.WriteString(indentString(dm.root.printNode(), 4))
	sb.WriteString("\n")

	return strings.TrimSpace(sb.String())
}
