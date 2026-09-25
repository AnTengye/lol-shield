//go:build !windows

package admin

// 非 Windows 仅用于 mock 开发和离线记录，不尝试 Windows 提权。
func MustRunWithAdmin(bool) {}

func RequireElevated() error { return nil }

func WatchParent(uint32, <-chan struct{}, func()) error { return nil }
