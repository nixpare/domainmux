package domainmux

import (
	"slices"
	"strings"
)

type expr interface {
	decrementPath() bool
	selectF(ctx *Context) bool
	selectFopt(ctx *Context) bool
	setArgsF(ctx *Context)
	setArgsFopt(ctx *Context)
}

type starExpr struct{
	index int
}

func (s *starExpr) decrementPath() bool {
	return true
}

func (s *starExpr) selectF(ctx *Context) bool {
	return len(ctx.path) > s.index
}

func (s *starExpr) selectFopt(ctx *Context) bool {
	return true
}

func (s *starExpr) setArgsF(ctx *Context) {}

func (s *starExpr) setArgsFopt(ctx *Context) {}

type paramExpr struct {
	key   string
	index int
}

func (p *paramExpr) decrementPath() bool {
	return true
}

func (p *paramExpr) selectF(ctx *Context) bool {
	return len(ctx.path) == p.index + 1
}

func (p *paramExpr) selectFopt(ctx *Context) bool {
	return len(ctx.path) == p.index + 1 || len(ctx.path) == p.index
}

func (p *paramExpr) setArgsF(ctx *Context) {
	if p.key == "_" {
		return
	}

	if len(ctx.path) > p.index {
		ctx.args[p.key] = ctx.path[p.index]
	} else {
		delete(ctx.args, p.key)
	}
}

func (p *paramExpr) setArgsFopt(ctx *Context) {
	if p.key == "_" {
		return
	}

	if len(ctx.path) > p.index {
		ctx.args[p.key] = ctx.path[p.index]
	} else if len(ctx.path) == p.index {
		ctx.args[p.key] = ""
	} else {
		delete(ctx.args, p.key)
	}
}

type variadicExpr struct {
	key   string
	index int
}

func (v *variadicExpr) decrementPath() bool {
	return true
}

func (v *variadicExpr) selectF(ctx *Context) bool {
	return len(ctx.path) != 0
}

func (v *variadicExpr) selectFopt(ctx *Context) bool {
	return true
}

func (v *variadicExpr) setArgsF(ctx *Context) {
	if v.key == "_" {
		return
	}

	length := len(ctx.path) - v.index
	if length <= 0 {
		delete(ctx.args, v.key)
		return
	}

	pathCopy := make([]string, length)
	copy(pathCopy, ctx.path[len(ctx.path)-length:])
	slices.Reverse(pathCopy)

	ctx.args[v.key] = strings.Join(pathCopy, ".")
}

func (v *variadicExpr) setArgsFopt(ctx *Context) {
	if v.key == "_" {
		return
	}

	length := len(ctx.path) - v.index
	if length < 0 {
		delete(ctx.args, v.key)
		return
	} else if length == 0 {
		ctx.args[v.key] = ""
		return
	}

	pathCopy := make([]string, length)
	copy(pathCopy, ctx.path[len(ctx.path)-length:])
	slices.Reverse(pathCopy)

	ctx.args[v.key] = strings.Join(pathCopy, ".")
}

type literalExpr struct {
	name         string
	index        int
	isAfterParam bool
	isLast       bool
}

func (l *literalExpr) decrementPath() bool {
	return l.isAfterParam
}

func (l *literalExpr) selectF(ctx *Context) bool {
	if !l.isAfterParam {
		if l.isLast {
			return len(ctx.path) == 0
		} else {
			return true
		}
	}

	if len(ctx.path) <= l.index {
		return false
	}

	if ctx.path[l.index] != l.name {
		return false
	}

	if !l.isLast {
		return true
	}

	return len(ctx.path) == l.index+1
}

func (l *literalExpr) selectFopt(ctx *Context) bool {
	panic("literalExpr.selectFopt should not be called")
}

func (l *literalExpr) setArgsF(ctx *Context) {}

func (l *literalExpr) setArgsFopt(ctx *Context) {
	panic("literalExpr.setArgsFopt should not be called")
}
