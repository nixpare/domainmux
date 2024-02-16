package domainmux

import (
	"net"
	"strconv"
	"strings"
)

// indentString takes a string and indents every line with
// the provided number of single spaces
func indentString(s string, n int) string {
	split := strings.Split(s, "\n")
	var res string

	for _, line := range split {
		for i := 0; i < n; i++ {
			res += " "
		}
		res += line + "\n"
	}

	return strings.TrimRight(res, " \n")
}

func SplitAddrPort(host string) string {
	sHost, _, err := net.SplitHostPort(host)
	if err != nil {
		sHost = host
	}
	return sHost
}

func SplitDomainSubdomain(host string) (domain string, subdomain string) {
	host = SplitAddrPort(host)

	split := strings.Split(host, ".")
	splitL := len(split)

	if splitL == 1 {
		domain = host
	} else {
		if _, err := strconv.Atoi(split[splitL-1]); err == nil {
			domain = host
		} else if strings.HasSuffix(host, "localhost") {
			domain = "localhost"
			subdomain = strings.Join(split[:splitL-1], ".")
		} else {
			domain = split[splitL-2] + "." + split[splitL-1]
			subdomain = strings.Join(split[:splitL-2], ".")
		}
	}
	
	return
}