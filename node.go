package domainmux

import (
	"fmt"
	"strings"
)

type node struct {
	path      string
	mws       []*handler
	childs    map[string]*node
	childkeys []string
}

func (n *node) createNode(path []string, leafMWs []*handler) {
	if len(path) == 0 {
		n.mws = append(n.mws, leafMWs...)
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
	next.createNode(path, leafMWs)
}

func (n *node) printNode() string {
	sb := strings.Builder{}
	sb.WriteString("Name: ")
	sb.WriteString(n.path)
	sb.WriteString(" - Middlewares: ")
	sb.WriteString(fmt.Sprint(len(n.mws)))
	sb.WriteString(" - Childs (")
	sb.WriteString(fmt.Sprint(len(n.childkeys)))
	sb.WriteString("): \n")

	for _, key := range n.childkeys {
		sb.WriteString(indentString(n.childs[key].printNode(), 4))
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}
