package domainmux

import (
	"net/http"
	"sync"
)

func isLocalDefault(remoteAddr string) bool {
	return remoteAddr == "localhost" || remoteAddr == "127.0.0.1" || remoteAddr == "::1"
}

func (dm *DomainMux) RedirectIfLocal(isLocal func(remoteAddr string) bool, rerun bool) error {
	lcm := &localClientManager{
		m: new(sync.RWMutex),
		clients: make(map[string]offlineClient),
		rerun: rerun,
	}

	if isLocal == nil {
		isLocal = func(remoteAddr string) bool { return false }
	}

	return dm.Serve("*", func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		remoteAddr := SplitAddrPort(r.RemoteAddr)
		if isLocal(remoteAddr) || isLocalDefault(remoteAddr) {
			lcm.handlerLocalQuery(ctx, r)
		}
	})
}

type offlineClient struct {
	domain    string
	subdomain string
}

type localClientManager struct {
	m *sync.RWMutex
	clients map[string]offlineClient
	rerun bool
}

func (lcm *localClientManager) handlerLocalQuery(ctx *Context, r *http.Request) {
	remoteAddr := SplitAddrPort(r.RemoteAddr)
	reqDomain, reqSubdomain := SplitDomainSubdomain(ctx.Host())
	domain, subdomain := reqDomain, reqSubdomain

	query := r.URL.Query()

	lcm.m.RLock()
	conf, ok := lcm.clients[remoteAddr]
	lcm.m.RUnlock()

	if ok {
		domain, subdomain = conf.domain, conf.subdomain
	}

	var updated bool

	if query.Has("domain") {
		updated = true
		domain = query.Get("domain")
	}

	if query.Has("subdomain") {
		updated = true
		subdomain = query.Get("subdomain")
	}

	if (domain == "" || domain == reqDomain) && (subdomain == "" || subdomain == reqSubdomain) {
		return
	}

	if subdomain == "" {
		subdomain = reqSubdomain
	}
	host := subdomain

	if host != "" {
		host += "."
	}

	if domain == "" {
		domain = reqDomain
	}
	host += domain

	if updated {
		lcm.m.Lock()
		lcm.clients[remoteAddr] = offlineClient{ domain, subdomain }
		lcm.m.Unlock()
	}

	ctx.Redirect(host, lcm.rerun)
}