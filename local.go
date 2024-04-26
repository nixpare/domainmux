package domainmux

import (
	"net/http"
	"sync"
)

type localClientManager struct {
	m *sync.RWMutex
	clients map[string]string
	rerun bool
	isLocal func(remoteAddr string) bool
}

func (dm *DomainMux) RedirectIfLocal(isLocal func(remoteAddr string) bool, rerun bool) {
	if isLocal == nil {
		isLocal = func(remoteAddr string) bool { return false }
	}

	lcm := &localClientManager{
		m: new(sync.RWMutex),
		clients: make(map[string]string),
		rerun: rerun,
		isLocal: isLocal,
	}

	dm.Serve("*", nil, lcm)
}

func (lcm *localClientManager) ServeDomainMux(ctx *Context, w http.ResponseWriter, r *http.Request) {
	remoteAddr := SplitAddrPort(r.RemoteAddr)
	if !lcm.isLocal(remoteAddr) && !isLocalDefault(remoteAddr) {
		return
	}

	domain := ctx.Host()
	query := r.URL.Query()

	lcm.m.RLock()
	savedDomain, ok := lcm.clients[remoteAddr]
	lcm.m.RUnlock()

	if ok {
		domain = savedDomain
	}

	var updated bool
	if query.Has("domain") {
		updated = true
		domain = query.Get("domain")
	}

	if domain == "" || domain == ctx.Host() {
		return
	}

	if !updated {
		w.Header().Set("Cache-Control", "no-cache")
		ctx.Redirect(domain, lcm.rerun)
		return
	}
		
	lcm.m.Lock()
	lcm.clients[remoteAddr] = domain
	lcm.m.Unlock()

	ctx.SetServeCalled()

	query.Del("domain")
	path := r.URL.Path
	if encQuery := query.Encode(); encQuery != "" {
		path += "?" + encQuery
	}

	http.Redirect(w, r, path, http.StatusTemporaryRedirect)
}

func isLocalDefault(remoteAddr string) bool {
	return remoteAddr == "localhost" || remoteAddr == "127.0.0.1" || remoteAddr == "::1"
}
