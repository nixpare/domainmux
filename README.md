# DomainMux

DomainMux is a powerful and flexible HTTP routing library for Go, designed to handle requests based on domain patterns. It provides a simple and intuitive API for building complex routing rules, making it easy to manage middleware, serve handlers, and extract parameters from domain names.

## Features

- **Domain-Based Routing:** Route requests to different handlers based on the requested domain or subdomain.
- **Middleware Support:** Easily apply middleware to specific domain patterns for granular control over request processing.
- **Parameter Extraction:** Extract dynamic values from domain names and use them in your handlers.
- **Flexible Pattern Matching:** Use wildcards (`*`), variadic parameters (`...`), and optional parameters (`?`) to define complex routing rules.
- **Standard `http.Handler` Compatibility:** Seamlessly integrate with existing `http.Handler` and `http.HandlerFunc` implementations.
- **Context Management:** A rich `Context` object provides access to request-specific data, including parameters and the `DomainMux` instance.

## Getting Started

To install DomainMux, use `go get`:

```bash
go get github.com/nixpare/domainmux
```

## Example

Here's a simple example demonstrating how to use DomainMux to route requests for different subdomains:

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/nixpare/domainmux"
)

func main() {
	mux := domainmux.NewDomainMux()

	// Serve a handler for the root domain
	mux.Serve("example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the root domain!")
	}))

	// Serve a handler for a specific subdomain
	mux.Serve("api.example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This is the API endpoint.")
	}))

	// Use a wildcard to handle all other subdomains
	mux.Serve("*.example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the %s subdomain!", ctx.Param("*"))
	}))

	// Start the server
	http.ListenAndServe(":8080", mux)
}
```

## Advanced Examples

DomainMux supports more complex routing scenarios through named parameters, variadic parameters, and optional parameters.

### Named Parameters (`$param`)

You can capture parts of the domain name into variables. The special name `$_` can be used as a wildcard if you need to match a part but don't need to capture its value.

```go
// Matches domains like "api.example.com" or "user.example.com"
// and captures the subdomain in the "service" parameter.
mux.Serve("$service.example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
    service := ctx.Param("service") // "api", "user", etc.
    fmt.Fprintf(w, "Welcome to the %s service!", service)
}))

// Matches "john.users.example.com" but only captures "john" and "com".
mux.Serve("$name.users.$_.$tld", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
    name := ctx.Param("name")     // "john"
    tld := ctx.Param("tld")       // "com"
    fmt.Fprintf(w, "Hello, %s from a .%s domain!", name, tld)
}))
```

### Variadic Parameters (`...param`)

Use `...` to match multiple subdomains at the beginning of a pattern. This is useful for catch-all scenarios.

```go
// Matches "a.b.c.example.com" and captures "a.b.c"
// in the "subdomains" parameter.
mux.Serve("...subdomains.example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
    subdomains := ctx.Param("subdomains") // "a.b.c"
    fmt.Fprintf(w, "You are on the following subdomains: %s", subdomains)
}))
```

### Optional Parameters (`?`)

Use `?` to make a parameter optional. This allows a pattern to match a domain with or without that specific part. The optional part must be the leftmost element in the pattern.

```go
// Matches both "example.com" and "staging.example.com"
mux.Serve("$env?.example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
    env := ctx.Param("env")
    if !ctx.HasParam("env") {
        fmt.Fprintln(w, "Welcome to the production environment!")
    } else {
        fmt.Fprintf(w, "Welcome to the %s environment!", env)
    }
}))

// This also works with variadic parameters.
// Matches "example.com", "a.example.com", "a.b.example.com", etc.
mux.Serve("...subs?.example.com", domainmux.HandlerFunc(func(ctx *domainmux.Context, w http.ResponseWriter, r *http.Request) {
    subs := ctx.Param("subs")
    if !ctx.HasParam("subs") {
        fmt.Fprintln(w, "Root domain!")
    } else {
        fmt.Fprintf(w, "Subdomains: %s", subs)
    }
}))
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
