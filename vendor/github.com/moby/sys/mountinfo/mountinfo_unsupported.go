//go:build !windows && !linux && !freebsd && !openbsd && !darwin
// +build !windows,!linux,!freebsd,!openbsd,!darwin

package mountinfo

import (
	"fmt"
	"runtime"
)

var errNotImplemented = fmt.Errorf("not implemented on %s/%s", "linux", runtime.GOARCH)

func parseMountTable(_ FilterFunc) ([]*Info, error) {
	return nil, errNotImplemented
}

func mounted(path string) (bool, error) {
	return false, errNotImplemented
}
