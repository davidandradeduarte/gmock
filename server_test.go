package gmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// default port
	want := &Server{
		port: defaultPort,
	}
	got := NewServer()
	assert.Equal(t, want.port, got.port)
}

func TestNewServerWithConfig(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		{
			name: "default port",
			args: args{
				config: Config{},
			},
			want: &Server{
				port:  defaultPort,
				stubs: make(stubHashMap),
			},
		},
		{
			name: "custom port",
			args: args{
				config: Config{
					Port: 9999,
				},
			},
			want: &Server{
				port:  9999,
				stubs: make(stubHashMap),
			},
		},
		{
			name: "custom port and stubs",
			args: args{
				config: Config{
					Port: 9999,
					Stubs: []*Stub{
						{
							Request: StubRequest{
								Method: "GET",
								Path:   "/test",
							},
							Response: StubResponse{
								StatusCode: 200,
								Body:       "{}",
							},
						},
					},
				},
			},
			want: &Server{
				port: 9999,
				stubs: stubHashMap{
					hash("cf31ce76c1b07cfaecef83c3feb6c04dd2c98f4f2aa33375c7b7a7d2a90424da"): &Stub{
						Request: StubRequest{
							Method: "GET",
							Path:   "/test",
						},
						Response: StubResponse{
							StatusCode: 200,
							Body:       "{}",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want.port, NewServerWithConfig(tt.args.config).port)
			assert.Equal(t, tt.want.stubs, NewServerWithConfig(tt.args.config).stubs)
		})
	}
}
