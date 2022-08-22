package gmock // nolint:golint

import "net/http"

// map with the current valid http methods
// https://github.com/golang/go/blob/master/src/net/http/method.go
var httpMethods = map[string]string{
	http.MethodGet:     http.MethodGet,
	http.MethodHead:    http.MethodHead,
	http.MethodPost:    http.MethodPost,
	http.MethodPut:     http.MethodPut,
	http.MethodPatch:   http.MethodPatch,
	http.MethodDelete:  http.MethodDelete,
	http.MethodConnect: http.MethodConnect,
	http.MethodOptions: http.MethodOptions,
	http.MethodTrace:   http.MethodTrace,
}
