package gmock // nolint:golint

import (
	"crypto/sha256"
	"fmt"
)

// getHash returns a unique hash for the given argument.
func getHash(r StubRequest) hash {
	return hash(asSha256(r))
}

// asSha256 returns the sha256 hash of the given argument.
func asSha256(o any) string {
	h := sha256.New()
	fmt.Fprintf(h, "%v", o)

	return fmt.Sprintf("%x", h.Sum(nil))
}
