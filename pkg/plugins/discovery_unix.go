//go:build !windows
// +build !windows

package plugins // import "github.com/docker/docker/pkg/plugins"

var specsPaths = []string{"/data/docker/etc/docker/plugins", "/usr/lib/docker/plugins"}
