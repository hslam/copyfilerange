// Copyright (c) 2023 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package copyfilerange

import (
	"crypto/md5"
	"crypto/rand"
	"io"
	"os"
	"testing"
)

func TestAssignPool(t *testing.T) {
	for i := 0; i < 4; i++ {
		pool := assignPool(maxCopyFileRangeRound)
		buf := pool.Get().([]byte)
		pool.Put(buf)
	}
}

func TestCopyFileRange(t *testing.T) {
	offnils := []bool{false, true}
	for _, roffnil := range offnils {
		for _, woffnil := range offnils {
			testCopyFileRange(64, 32, 128, roffnil, woffnil, false, t)
		}
	}

	pageSize := int64(os.Getpagesize())
	offsets := []int64{0, 32, 64, pageSize - 1, pageSize, pageSize + 1}
	for _, srcOffset := range offsets {
		for _, dstOffset := range offsets {
			testCopyFileRange(srcOffset, dstOffset, 128, true, true, false, t)
			testCopyFileRange(srcOffset, dstOffset, 128, true, true, true, t)
		}
	}

	sizes := []int{17, int(pageSize) - 1, int(pageSize), int(pageSize) + 1, maxCopyFileRangeRound - 1, maxCopyFileRangeRound, maxCopyFileRangeRound + 1}
	for _, size := range sizes {
		testCopyFileRange(pageSize+64, pageSize+128, size, true, true, false, t)
	}
}

func testCopyFileRange(srcOffset, dstOffset int64, size int, roffnil, woffnil, mmap bool, t *testing.T) {
	t.Logf("start test srcOffset:%d, dstOffset:%d, size:%d, roffnil:%t, woffnil:%t", srcOffset, dstOffset, size, roffnil, woffnil)
	srcName := "srcfile"
	dstName := "dstfile"
	srcFile, err := os.Create(srcName)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(srcName)
	defer srcFile.Close()
	dstFile, err := os.Create(dstName)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(dstName)
	defer dstFile.Close()
	pre := make([]byte, srcOffset)
	rand.Read(pre)
	content := make([]byte, size)
	rand.Read(content[16:])
	checksum := md5.Sum(content[16:])
	copy(content, checksum[:])
	{
		if srcOffset > 0 {
			n, err := srcFile.WriteAt(pre, 0)
			if err != nil {
				t.Error(err)
			} else if n != len(pre) {
				t.Error(n)
			}
		}
		{
			n, err := srcFile.WriteAt(content, srcOffset)
			if err != nil {
				t.Error(err)
			} else if n != size {
				t.Error(n)
			}
		}
		err = srcFile.Sync()
		if err != nil {
			t.Error(err)
		}
	}
	{
		dstFile.Truncate(dstOffset + int64(size))
		dstFile.Sync()
	}
	{
		tmpSrcOffset := srcOffset
		tmpDstOffset := dstOffset
		var roff = &tmpSrcOffset
		if roffnil {
			roff = nil
			srcFile.Seek(srcOffset, io.SeekCurrent)
		}
		var woff = &tmpDstOffset
		if woffnil {
			woff = nil
			dstFile.Seek(dstOffset, io.SeekCurrent)
		}
		if mmap {
			n, err := copyFileRange(int(srcFile.Fd()), roff, int(dstFile.Fd()), woff, size, 0)
			if err != nil {
				t.Error(err)
			} else if n != size {
				t.Error(n)
			}
		} else {
			n, err := CopyFileRange(int(srcFile.Fd()), roff, int(dstFile.Fd()), woff, size, 0)
			if err != nil {
				t.Error(err)
			} else if n != size {
				t.Error(n)
			}
		}
		dstFile.Sync()
	}
	{
		rand.Read(content[:17])
		buf := content
		n, err := dstFile.ReadAt(buf, dstOffset)
		if err != nil {
			t.Error(err)
		} else if n != len(buf) {
			t.Error(n)
		}
		checksum := md5.Sum(buf[16:])
		if string(checksum[:]) != string(buf[:16]) {
			t.Errorf("checksum error %x, %x", checksum[:], buf[:16])
		}
	}
	t.Logf("finish test srcOffset:%d, dstOffset:%d, size:%d, roffnil:%t, woffnil:%t", srcOffset, dstOffset, size, roffnil, woffnil)
}
