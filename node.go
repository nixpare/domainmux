package domainmux

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type nodeMap struct {
	m    map[string]*Node
	keys []string
}

func (pool *nodeMap) createNode(path []string, index int, leafMWs []Handler) {
	n := pool.m[path[index]]
	if n == nil {
		n = &Node{
			path:   append([]string{}, path[:index+1]...),
			childs: &nodeMap{m: make(map[string]*Node)},
		}

		pool.m[path[index]] = n
		pool.keys = append(pool.keys, path[index])
	}

	if index == len(path)-1 {
		n.mws = append(n.mws, leafMWs...)
		return
	}

	n.childs.createNode(path, index+1, leafMWs)
}

type Node struct {
	path   []string
	mws    []Handler
	childs *nodeMap
}

func (n *Node) name() string {
	return strings.Join(n.path, "->")
}

func (n *Node) printNode() string {
	sb := strings.Builder{}
	sb.WriteString("Name: ")
	sb.WriteString(n.name())
	sb.WriteString(" - Middlewares: ")
	sb.WriteString(fmt.Sprint(len(n.mws)))
	sb.WriteString(" - Childs (")
	sb.WriteString(fmt.Sprint(len(n.childs.keys)))
	sb.WriteString("): \n")

	for _, key := range n.childs.keys {
		sb.WriteString(indentString(n.childs.m[key].printNode(), 4))
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}

type fetchRes struct {
	mws []Handler
	childs []*fetchRes
}

func (n *Node) fetchHandlers(path []string, index int) (res *fetchRes) {
	res = new(fetchRes)

	// if the current path is variadic
	if strings.HasPrefix(n.path[index], "...") {
		// variadic domains does not have childs, so returns only
		// its handlers
		res.mws = n.argsHandler(path)
		return
	}
	// else the current path is not variadic

	// if the current parts do not match, this path is not corrent, so return
	if path[index] != n.path[index] && !strings.HasPrefix(n.path[index], ":") {
		return
	}

	// if we are at the last path element and the two paths have the same length
	// appends it's handlers 
	if index == len(path)-1 && len(path) == len(n.path) {
		res.mws = n.argsHandler(path)
		return
	}

	// else if there are more segments to match, append the childs
	if len(path) > len(n.path) {
		for _, key := range n.childs.keys {
			value := n.childs.m[key]
			
			res.childs = append(res.childs, value.fetchHandlers(path, index+1))
		}
	}

	return res
}

func (n *Node) argsHandler(path []string) []Handler {
	args := make(map[string]string)
	for i := range n.path {
		switch {
		case strings.HasPrefix(n.path[i], "..."):
			key := strings.TrimLeft(n.path[i], ".")
			value := append([]string{}, path[i:]...)
			slices.Reverse(value)
			args[key] = strings.Join(value, ".")
		case strings.HasPrefix(n.path[i], ":"):
			key := strings.TrimLeft(n.path[i], ":")
			args[key] = path[i]
		}
	}

	var res []Handler
	for _, h := range n.mws {
		res = append(res, func(ctx *Context, w http.ResponseWriter, r *http.Request) {
			for key, value := range args {
				if key != "_" {
					ctx.args[key] = value
				}
			}
			h(ctx, w, r)
		})
	}

	return res
}
