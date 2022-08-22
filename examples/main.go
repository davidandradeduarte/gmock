// nolint
package main

import (
	"net/http"
	"net/url"
	"time"

	"github.com/davidandradeduarte/gmock"
)

var stub = &gmock.Stub{
	Request:  gmock.StubRequest{Method: http.MethodGet, Path: "/test", Query: url.Values{"id": []string{"1"}}, Body: `{"from":"json"}`},
	Response: gmock.StubResponse{StatusCode: http.StatusOK, Body: `{"to":"json"}`, Headers: http.Header{"X-Test": []string{"test"}}},
}

func main() {
	// you can also start the server without storing the server instance
	// (it will stop when the main function exits)
	// e.g
	// gmock.NewServer().Start()
	// or you can even start it in a goroutine
	// go gmock.NewServerWithConfig(gmock.Config{Port: 8888}).Start()

	sv := gmock.NewServer()
	sv.
		WithStubsFrom("pkg/gmock/examples").
		WithStubs(stub, stub).
		WithStubs([]*gmock.Stub{stub, stub}...).
		WithPort(8888).
		Start()

	sv.ClearStubs()
	sv.AddStub(stub)
	sv.AddStubs(stub, stub)
	sv.AddStubs([]*gmock.Stub{stub, stub}...)

	time.Sleep(60 * time.Second)
	if err := sv.Stop(); err != nil {
		panic(err)
	}

	sv = gmock.NewServerWithConfig(gmock.Config{
		Port:  8888,
		Stubs: []*gmock.Stub{stub, stub},
	})
	sv.Start()

	time.Sleep(60 * time.Second)
	if err := sv.Stop(); err != nil {
		panic(err)
	}
}
