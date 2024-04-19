package domainmux

import (
	"net/http"
	"strings"
)

func (dm *DomainMux) Aliases(host string, rerun bool, matchF func(host string) bool, aliases ...string) error {
	if matchF == nil {
		matchF = func(host string) bool { return false }
	}

	return dm.Serve("*", func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		for _, a := range aliases {
			if index := strings.LastIndex(ctx.Host(), a); index != -1 {
				redirect := ctx.Host()[:index] + host
				ctx.Redirect(redirect, rerun)
				return
			}
		}

		if matchF(ctx.Host()) {
			ctx.Redirect(host, rerun)
		}
	})
}