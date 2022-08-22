package gmock

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubRequest_Validate(t *testing.T) {
	type fields struct {
		Method  string
		Path    string
		Query   url.Values
		Body    any
		Headers map[string][]string
	}
	tests := []struct {
		name   string
		fields fields
		want   []error
	}{
		{
			name: "valid request",
			fields: fields{
				Method: http.MethodGet,
				Path:   "/test",
			},
			want: nil,
		},
		{
			name: "required method",
			fields: fields{
				Path: "/test",
			},
			want: []error{&errRequiredMethod{}},
		},
		{
			name: "invalid method",
			fields: fields{
				Method: "INVALID",
				Path:   "/test",
			},
			want: []error{&errInvalidMethod{method: "INVALID"}},
		},
		{
			name: "invalid path",
			fields: fields{
				Method: http.MethodGet,
				Path:   "",
			},
			want: []error{&errRequiredPath{}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := &StubRequest{
				Method: tt.fields.Method,
				Path:   tt.fields.Path,
				Query:  tt.fields.Query,
				Body:   tt.fields.Body,
			}
			assert.Equal(t, tt.want, r.Validate())
		})
	}
}

func TestStubRequest_Sanitize(t *testing.T) {
	r := &StubRequest{
		Path: "test",
	}
	r.Sanitize()
	assert.Equal(t, "/test", r.Path)
	r = &StubRequest{
		Path: "/test",
	}
	r.Sanitize()
	assert.Equal(t, "/test", r.Path)
}

func TestStubResponse_Validate(t *testing.T) {
	type fields struct {
		StatusCode int
		Headers    map[string][]string
		Body       any
	}
	tests := []struct {
		name   string
		fields fields
		want   []error
	}{
		{
			name: "valid response",
			fields: fields{
				StatusCode: 200,
			},
			want: nil,
		},
		{
			name: "invalid status code",
			fields: fields{
				StatusCode: -1,
			},
			want: []error{&errInvalidStatusCode{statusCode: -1}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := &StubResponse{
				StatusCode: tt.fields.StatusCode,
				Headers:    tt.fields.Headers,
				Body:       tt.fields.Body,
			}
			assert.Equal(t, tt.want, r.Validate())
		})
	}
}

func TestStub_validationErrors(t *testing.T) {
	// valid
	r := &Stub{
		Request: StubRequest{
			Method: http.MethodGet,
			Path:   "/test",
		},
		Response: StubResponse{
			StatusCode: http.StatusOK,
		},
	}
	assert.Nil(t, r.validationErrors())

	// invalid
	r = &Stub{
		Request: StubRequest{
			Method: "invalid",
		},
	}
	assert.GreaterOrEqual(t, len(r.validationErrors()), 1)
}

func Test_getStubFromBytes(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Stub
		wantErr bool
	}{
		{
			name: "empty stub",
			args: args{
				b: []byte(""),
			},
			want: nil,
		},
		{
			name: "empty json stub",
			args: args{
				b: []byte(`{}`),
			},
			want: &Stub{},
		},
		{
			name: "valid json stub",
			args: args{
				b: []byte(`{
					"request": {
						"method": "GET",
						"path": "/test"
					},
					"response": {
						"status_code": 200
					}
				}`),
			},
			want: &Stub{
				Request: StubRequest{
					Method: http.MethodGet,
					Path:   "/test",
				},
				Response: StubResponse{
					StatusCode: http.StatusOK,
				},
			},
		},
		{
			name: "valid yaml stub",
			args: args{
				b: []byte(`
request:
  method: GET
  path: "/test"
response:
  status_code: 200
`),
			},
			want: &Stub{
				Request: StubRequest{
					Method: http.MethodGet,
					Path:   "/test",
				},
				Response: StubResponse{
					StatusCode: http.StatusOK,
				},
			},
		},
		{
			name: "invalid stub",
			args: args{
				b: []byte(`invalid`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStubFromBytes(tt.args.b)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
			assert.NoError(t, err)
		})
	}
}
