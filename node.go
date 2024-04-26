package domainmux

import (
	"fmt"
	"strings"
)

type node struct {
	path      string
	handlers  []*nodeHandler
	childs    map[string]*node
	childkeys []string
}

func (n *node) createNode(path []string, handler *nodeHandler) {
	if len(path) == 0 {
		n.handlers = append(n.handlers, handler)
		return
	}

	next := n.childs[path[0]]
	if next == nil {
		next = &node{
			path:   path[0],
			childs: make(map[string]*node),
		}

		n.childkeys = append(n.childkeys, path[0])
		n.childs[path[0]] = next
	}

	path = path[1:]
	next.createNode(path, handler)
}

func (n *node) printNode() string {
	sb := strings.Builder{}
	sb.WriteString("Name: ")
	sb.WriteString(n.path)

	if n.handlers != nil {
		var (
			mwCount int
			serveFound bool
		)
		for _, h := range n.handlers {
			mwCount += len(h.mws)
			serveFound = serveFound || h.serveHandler != nil
		}

		sb.WriteString(" - Middlewares: ")
		sb.WriteString(fmt.Sprint(mwCount))
		sb.WriteString(" - ServeFunc: ")
		sb.WriteString(fmt.Sprint(serveFound))
	}
	
	sb.WriteString(" - Childs (")
	sb.WriteString(fmt.Sprint(len(n.childkeys)))
	sb.WriteString("): \n")

	for _, key := range n.childkeys {
		sb.WriteString(indentString(n.childs[key].printNode(), 4))
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}
