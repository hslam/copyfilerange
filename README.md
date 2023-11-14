# copyfilerange
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hslam/copyfilerange)](https://pkg.go.dev/github.com/hslam/copyfilerange)
[![Build Status](https://github.com/hslam/copyfilerange/workflows/build/badge.svg)](https://github.com/hslam/copyfilerange/actions)
[![codecov](https://codecov.io/gh/hslam/copyfilerange/branch/master/graph/badge.svg)](https://codecov.io/gh/hslam/copyfilerange)
[![Go Report Card](https://goreportcard.com/badge/github.com/hslam/copyfilerange)](https://goreportcard.com/report/github.com/hslam/copyfilerange)
[![LICENSE](https://img.shields.io/github/license/hslam/copyfilerange.svg?style=flat-square)](https://github.com/hslam/copyfilerange/blob/master/LICENSE)

Package copyfilerange wraps the copy_file_range system call.

## Get started

### Install
```
go get github.com/hslam/copyfilerange
```
### Import
```
import "github.com/hslam/copyfilerange"
```
### Usage
#### Example
```go
package main

import (
	"fmt"
	"github.com/hslam/copyfilerange"
	"os"
)

func main() {
	srcName, dstName := "srcFile", "dstFile"
	srcFile, _ := os.Create(srcName)
	defer os.Remove(srcName)
	defer srcFile.Close()
	dstFile, _ := os.Create(dstName)
	defer os.Remove(dstName)
	defer dstFile.Close()

	content := []byte("Hello world")
	srcOffset, dstOffset := int64(64), int64(32)
	srcFile.Truncate(srcOffset)
	srcFile.WriteAt(content, srcOffset)
	dstFile.Truncate(dstOffset + int64(len(content)))

	roff, woff := srcOffset, dstOffset
	copyfilerange.CopyFileRange(int(srcFile.Fd()), &roff, int(dstFile.Fd()), &woff, len(content), 0)

	buf := make([]byte, len(content))
	dstFile.ReadAt(buf, dstOffset)
	fmt.Println(string(buf))
}
```

### Output
```
Hello world
```

### License
This package is licensed under a MIT license (Copyright (c) 2023 Meng Huang)


### Author
copyfilerange was written by Meng Huang.
