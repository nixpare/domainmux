package domainmux

import (
	"net/http"
	"testing"
)

func setParam(ctx *Context, query string, testMap map[string]string) {
	if ctx.HasValue(query) {
		testMap[query] = ctx.Value(query)
	}
}

func TestMain(t *testing.T) {
	dm := NewDomainMux()

	testMap := make(map[string]string)

	dm.Middleware("*", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["*"] = "OK"
		
		ctx.Next(w, r)
		ctx.SetServeCalled()
	}))

	dm.Middleware("$sub.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["$sub.test.com"] = "OK"
		setParam(ctx, "sub", testMap)
	}))

	dm.Middleware("$_.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["$_.test.com"] = "OK"
	}))

	dm.Middleware("...subs.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["...subs.test.com"] = "OK"
		setParam(ctx, "subs", testMap)
	}))

	dm.Middleware("$opt?.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["$opt?.test.com"] = "OK"
		setParam(ctx, "opt", testMap)
	}))

	dm.Middleware("...opts?.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["...opts?.test.com"] = "OK"
		setParam(ctx, "opts", testMap)
	}))

	dm.Serve("test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["test.com"] = "OK"
	}))

	dm.Serve("sub1.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["sub1.test.com"] = "OK"
	}))

	dm.Serve("more.sub1.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["more.sub1.test.com"] = "OK"
	}))

	dm.Serve("sub2.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		testMap["sub2.test.com"] = "OK"
	}))

	type testData struct {
		address string
		result map[string]string
	}

	requests := []testData{
		{ "test.com", map[string]string{
			"*": "OK", "test.com": "OK",
			"$opt?.test.com": "OK", "opt": "",
			"...opts?.test.com": "OK", "opts": "",
		} },
		{ "sub.test.com", map[string]string{
			"*": "OK",
			"$sub.test.com": "OK", "sub": "sub",
			"$_.test.com": "OK",
			"...subs.test.com": "OK", "subs": "sub",
			"$opt?.test.com": "OK", "opt": "sub",
			"...opts?.test.com": "OK", "opts": "sub",
		} },
		{ "sub1.test.com", map[string]string{
			"*": "OK",
			"$sub.test.com": "OK", "sub": "sub1",
			"$_.test.com": "OK",
			"...subs.test.com": "OK", "subs": "sub1",
			"$opt?.test.com": "OK", "opt": "sub1",
			"...opts?.test.com": "OK", "opts": "sub1",
			"sub1.test.com": "OK",
		} },
		{ "more.sub1.test.com", map[string]string{
			"*": "OK",
			"...subs.test.com": "OK", "subs": "more.sub1",
			"...opts?.test.com": "OK", "opts": "more.sub1",
			"more.sub1.test.com": "OK",
		} },
		{ "sub2.test.com", map[string]string{
			"*": "OK",
			"$sub.test.com": "OK", "sub": "sub2",
			"$_.test.com": "OK",
			"...subs.test.com": "OK", "subs": "sub2",
			"$opt?.test.com": "OK", "opt": "sub2",
			"...opts?.test.com": "OK", "opts": "sub2",
			"sub2.test.com": "OK",
		} },
		{ "example.com", map[string]string{
			"*": "OK",
		} },
		{ "sub.example.com", map[string]string{
			"*": "OK",
		} },
	}

	for _, req := range requests {
		clear(testMap)
		dm.ServeHTTP(nil, &http.Request{
			Host: req.address,
		})

		for key, expected := range req.result {
			value, ok := testMap[key]
			delete(testMap, key)
			if !ok {
				t.Errorf("on \"%s\" with param \"%s\" expected \"%s\" but not found", req.address, key, expected)
				continue
			}
			if value != expected {
				t.Errorf("on \"%s\" with param \"%s\" expected \"%s\" but found \"%s\"", req.address, key, expected, value)
			}
		}

		if len(testMap) > 0 {
			t.Errorf("on \"%s\" found leftovers %v", req.address, testMap)
		}
	}
}

type fakeWriter struct {}
func (fw *fakeWriter) Header() http.Header {
	return make(http.Header)
}
func (fw *fakeWriter) Write(b []byte) (int, error) {
	return 0, nil
}
func (fw *fakeWriter) WriteHeader(statuscode int) {}

func BenchmarkMain(b *testing.B) {
	dm := NewDomainMux()

	dm.Middleware("*", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		ctx.Next(w, r)
		ctx.SetServeCalled()
	}))

	dm.Middleware("$sub.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("$sub.test.com"))
	}))

	dm.Middleware("$_.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("$_.test.com"))
	}))

	dm.Middleware("...subs.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("...subs.test.com"))
	}))

	dm.Middleware("$opt?.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("$opt?.test.com"))
	}))

	dm.Middleware("...opts?.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("...opts?.test.com"))
	}))

	dm.Serve("test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test.com"))
	}))

	dm.Serve("sub1.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sub1.test.com"))
	}))

	dm.Serve("more.sub1.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("more.sub1.test.com"))
	}))

	dm.Serve("sub2.test.com", HandlerFunc(func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("sub2.test.com"))
	}))

	requests := []string{
		"test.com",
		"sub.test.com",
		"sub1.test.com",
		"more.sub1.test.com",
		"sub2.test.com",
		"example.com",
		"sub.example.com",
	}

	for range b.N {
		for _, req := range requests {
			dm.ServeHTTP(&fakeWriter{}, &http.Request{
				Host: req,
			})
		}
	}
}
