package middleware

import (
	"net/http"

	"github.com/nixpare/domainmux"
)

func Aliases(host string, matchF func(host string) bool, aliases ...string) domainmux.Handler {
	if matchF == nil {
		matchF = func(host string) bool { return false }
	}

	return func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
		for _, a := range aliases {
			if a == ctx.Host() {
				ctx.ChangeHost(host)
				return
			}
		}

		if matchF(ctx.Host()) {
			ctx.ChangeHost(host)
		}
	}
}