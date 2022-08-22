package gmock // nolint:golint

import (
	"encoding/json"
	"fmt"
	"net/url"

	"gopkg.in/yaml.v3"
)

// hash is a string that uniquely identifies a stub.
type hash string

// stubHashMap is a map of hash to stub.
type stubHashMap map[hash]*Stub

// Stub is a request and response pair that is used to match incoming requests.
type Stub struct {
	Request  StubRequest  `json:"request" yaml:"request"`
	Response StubResponse `json:"response" yaml:"response"`
}

// StubRequest is the request part of a Stub.
type StubRequest struct {
	Method string     `json:"method" yaml:"method"`
	Path   string     `json:"path" yaml:"path"`
	Query  url.Values `json:"query" yaml:"query"`
	Body   any        `json:"body" yaml:"body"`
}

// StubResponse is the response part of a Stub.
type StubResponse struct {
	StatusCode int                 `json:"status_code" yaml:"status_code"`
	Headers    map[string][]string `json:"headers" yaml:"headers"`
	Body       any                 `json:"body" yaml:"body"`
}

// Validate returns a list of validation errors for the current stub request.
func (r *StubRequest) Validate() []error {
	var errs []error
	if r.Method == "" {
		errs = append(errs, &errRequiredMethod{})
	} else if _, ok := httpMethods[r.Method]; !ok {
		errs = append(errs, &errInvalidMethod{method: r.Method})
	}
	if r.Path == "" {
		errs = append(errs, &errRequiredPath{})
	}
	return errs
}

// Sanitize sanitizes a stub request.
func (r *StubRequest) Sanitize() {
	if r.Path != "" && r.Path[0] != '/' {
		r.Path = "/" + r.Path
	}
}

// String returns a string representation of the stub request.
func (r *StubRequest) String() string {
	_url := r.Path
	first := true
	// fixme: write existing multiple values for the same key in query
	for k := range r.Query {
		if first {
			_url += "?"
			first = false
		} else {
			_url += "&"
		}
		_url += fmt.Sprintf("%s=%s", k, r.Query.Get(k))
	}
	return fmt.Sprintf(`%s %s %s`, r.Method, _url, r.Body)
}

// Validate returns a list of validation errors for the current stub response.
func (r *StubResponse) Validate() []error {
	var errs []error
	if r.StatusCode < 200 || r.StatusCode > 599 {
		errs = append(errs, &errInvalidStatusCode{statusCode: r.StatusCode})
	}
	return errs
}

// validationErrors returns a list of validation errors for the current stub.
func (r *Stub) validationErrors() []error {
	return append(r.Request.Validate(), r.Response.Validate()...)
}

// getStubFromBytest returns a stub from a byte array.
func getStubFromBytes(b []byte) (*Stub, error) {
	var stub *Stub
	if err := json.Unmarshal(b, stub); err != nil {
		if err := yaml.Unmarshal(b, &stub); err != nil {
			return nil, err
		}
	}
	return stub, nil
}
