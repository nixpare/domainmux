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

func (dm *DomainMux) ServeFunc(host string, serveFunc HandlerFunc, middlewareFuncs ...HandlerFunc) {
	mws := make([]Handler, 0, len(middlewareFuncs))
	for _, f := range middlewareFuncs {
		if f != nil {
			mws = append(mws, f)
		}
	}

	if serveFunc == nil {
		dm.Serve(host, nil, mws...)
	} else {
		dm.Serve(host, serveFunc, mws...)
	}
}

func (dm *DomainMux) Serve(host string, serveFunc Handler, middlewares ...Handler) {
	path, selectF, setArgsF, err := parseQuery(host)
	if err != nil {
		panic(fmt.Errorf("invalid query: %w", err))
	}

	h := &nodeHandler{
		serveHandler: serveFunc,
		mws: middlewares,
		selectF: selectF,
		setArgsF: setArgsF,
	}

	dm.root.createNode(path, h)
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
