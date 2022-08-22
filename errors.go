package gmock // nolint:golint

import "fmt"

type errRequiredMethod struct{}

func (*errRequiredMethod) Error() string {
	return "method is required"
}

type errInvalidMethod struct {
	method string
}

func (e *errInvalidMethod) Error() string {
	return fmt.Sprintf("method %s is not valid", e.method)
}

type errRequiredPath struct{}

func (*errRequiredPath) Error() string {
	return "path is required"
}

type errInvalidStatusCode struct {
	statusCode int
}

func (e *errInvalidStatusCode) Error() string {
	return fmt.Sprintf("status code %d is not valid", e.statusCode)
}
