package domainmux

import (
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

func (dm *DomainMux) Serve(host string, middlewares ...Handler) {
	path, selectF, setArgsF := parseQuery(host)

	handlers := make([]*handler, 0, len(middlewares))
	for _, mw := range middlewares {
		handlers = append(handlers, &handler{
			h: mw,
			selectF: selectF,
			setArgsF: setArgsF,
		})
	}

	dm.root.createNode(path, handlers)
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

	sb.WriteString(indentString(dm.root.printNode(), 4))
	sb.WriteString("\n")

	return strings.TrimSpace(sb.String())
}
