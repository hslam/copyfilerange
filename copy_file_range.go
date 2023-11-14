// Copyright (c) 2023 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package copyfilerange wraps the copy_file_range system call.
package copyfilerange

import (
	"github.com/hslam/mmap"
	"io"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	buffers = sync.Map{}
	assign  int32
)

func assignPool(size int) *sync.Pool {
	for {
		if p, ok := buffers.Load(size); ok {
			return p.(*sync.Pool)
		}
		if atomic.CompareAndSwapInt32(&assign, 0, 1) {
			var pool = &sync.Pool{New: func() interface{} {
				return make([]byte, size)
			}}
			buffers.Store(size, pool)
			atomic.StoreInt32(&assign, 0)
			return pool
		}
	}
}

func copyFileRange(rfd int, roff *int64, wfd int, woff *int64, length int, flags int) (written int, err error) {
	var remain = length
	var rpos = roff
	var wpos = woff
	var rerr, werr error
	if rpos == nil {
		var p int64
		p, rerr = syscall.Seek(rfd, 0, io.SeekCurrent)
		if rerr == nil {
			rpos = &p
		}
	}
	var pool *sync.Pool
	var buf []byte
	if rpos == nil {
		size := maxCopyFileRangeRound
		if remain < maxCopyFileRangeRound {
			size = remain
		}
		pool = assignPool(size)
		buf = pool.Get().([]byte)
		defer pool.Put(buf)
	}
	if wpos == nil {
		var p int64
		p, werr = syscall.Seek(wfd, 0, io.SeekCurrent)
		if werr == nil {
			wpos = &p
		}
	}
	for remain > 0 {
		size := maxCopyFileRangeRound
		if remain < maxCopyFileRangeRound {
			size = remain
		}
		var wn int
		var errno error
		if rpos != nil && wpos != nil {
			roffset := mmap.Offset(*rpos)
			rrel := *rpos - roffset
			var rb []byte
			rb, err = mmap.Open(rfd, roffset, int(rrel)+size, mmap.READ)
			if err != nil {
				return
			}
			woffset := mmap.Offset(*wpos)
			wrel := *wpos - woffset
			var wb []byte
			wb, err = mmap.Open(wfd, woffset, int(wrel)+size, mmap.WRITE)
			if err != nil {
				mmap.Munmap(rb)
				return
			}
			wn = copy(wb[wrel:wrel+int64(size)], rb[rrel:rrel+int64(size)])
			errno = mmap.Msync(wb)
			mmap.Munmap(wb)
			mmap.Munmap(rb)
		} else if rpos != nil && wpos == nil {
			roffset := mmap.Offset(*rpos)
			rrel := *rpos - roffset
			var rb []byte
			rb, err = mmap.Open(rfd, roffset, int(rrel)+size, mmap.READ)
			if err != nil {
				return
			}
			wn, errno = syscall.Write(wfd, rb[rrel:rrel+int64(size)])
			mmap.Munmap(rb)
		} else {
			if rpos != nil {
				syscall.Seek(rfd, *rpos, io.SeekStart)
			}
			var rn int
			rn, err = syscall.Read(rfd, buf)
			if err != nil {
				return
			}
			if wpos != nil {
				syscall.Seek(wfd, *wpos, io.SeekStart)
			}
			wn, err = syscall.Write(wfd, buf[:rn])
			if err != nil {
				return
			}
		}
		if wn > 0 {
			if rpos != nil {
				*rpos += int64(wn)
			}
			if wpos != nil {
				*wpos += int64(wn)
			}
			written += wn
			remain -= wn
		} else if (wn == 0 && errno == nil) || (errno != nil && errno != syscall.EAGAIN) {
			err = errno
			break
		}
	}
	return written, err
}
