package gmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_asSha256(t *testing.T) {
	type args struct {
		o any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test string",
			args: args{
				o: "test",
			},
			want: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			name: "empty string",
			args: args{
				o: "",
			},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "test struct",
			args: args{
				o: StubRequest{
					Method: "GET",
					Path:   "/test",
				},
			},
			want: "cf31ce76c1b07cfaecef83c3feb6c04dd2c98f4f2aa33375c7b7a7d2a90424da",
		},
		{
			name: "test empty struct",
			args: args{
				o: StubRequest{},
			},
			want: "aa2cf3db9ede69e284d3e1ba44c66bd54e7c6c23340fefcd2306cee3413e94f8",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, asSha256(tt.args.o))
		})
	}
}

func Test_getHash(t *testing.T) {
	want := hash("cf31ce76c1b07cfaecef83c3feb6c04dd2c98f4f2aa33375c7b7a7d2a90424da")
	got := getHash(StubRequest{Method: "GET", Path: "/test"})
	require.Equal(t, want, got)
}
