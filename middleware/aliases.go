package middleware

import (
	"net/http"
	"strings"

	"github.com/nixpare/domainmux"
)

func Aliases(host string, rerun bool, matchF func(host string) bool, aliases ...string) domainmux.Handler {
	if matchF == nil {
		matchF = func(host string) bool { return false }
	}

	return func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
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
	}
}