// Copyright (c) 2023 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

//go:build !darwin && !linux && !dragonfly && !freebsd && !netbsd && !openbsd
// +build !darwin,!linux,!dragonfly,!freebsd,!netbsd,!openbsd

package copyfilerange

const (
	maxCopyFileRangeRound int = 64 << 10
)

// CopyFileRange copies a range of data from one file to another.
func CopyFileRange(rfd int, roff *int64, wfd int, woff *int64, len int, flags int) (written int, err error) {
	return copyFileRange(rfd, roff, wfd, woff, len, flags)
}
