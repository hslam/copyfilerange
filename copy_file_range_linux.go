// Copyright (c) 2023 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

//go:build linux
// +build linux

package copyfilerange

import (
	"syscall"
	"unsafe"
)

const (
	maxCopyFileRangeRound = 1 << 30
)

// CopyFileRange copies a range of data from one file to another.
//
// The copy_file_range() system call performs an in-kernel copy
// between two file descriptors without the additional cost of
// transferring data from the kernel to user space and then back
// into the kernel.  It copies up to len bytes of data from the
// source file descriptor rfd to the target file descriptor
// wfd, overwriting any data that exists within the requested
// range of the target file.
//
// The following semantics apply for roff, and similar statements
// apply to woff:
//
// •  If roff is NULL, then bytes are read from rfd starting
// from the file offset, and the file offset is adjusted by the
// number of bytes copied.
//
// •  If roff is not NULL, then roff must point to a buffer that
// specifies the starting offset where bytes from rfd will be
// read.  The file offset of rfd is not changed, but roff is
// adjusted appropriately.
//
// rfd and wfd can refer to the same file.  If they refer to
// the same file, then the source and target ranges are not allowed
// to overlap.
//
// The flags argument is provided to allow for future extensions and
// currently must be set to 0.
func CopyFileRange(rfd int, roff *int64, wfd int, woff *int64, len int, flags int) (written int, err error) {
	var remain = len
	for remain > 0 {
		size := maxCopyFileRangeRound
		if int(remain) < maxCopyFileRangeRound {
			size = int(remain)
		}
		n, errno := syscallCopyFileRange(rfd, roff, wfd, woff, size, flags)
		if n > 0 {
			written += n
			remain -= n
		} else if (n == 0 && errno == nil) || (errno != nil && errno != syscall.EINTR) {
			err = errno
			break
		}
	}
	return
}

func syscallCopyFileRange(rfd int, roff *int64, wfd int, woff *int64, len int, flags int) (n int, err error) {
	r1, _, errno := syscall.Syscall6(copyFileRangeTrap,
		uintptr(rfd),
		uintptr(unsafe.Pointer(roff)),
		uintptr(wfd),
		uintptr(unsafe.Pointer(woff)),
		uintptr(len),
		uintptr(flags),
	)
	n = int(r1)
	if errno != 0 {
		err = errno
	}
	return
}
