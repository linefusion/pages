package s3

import (
	"io"
	"io/fs"
	"path"
	"time"
)

var (
	_ fs.File     = (*File)(nil)
	_ fs.FileInfo = (*FileInfo)(nil)
)

type File struct {
	io.ReadCloser
	stat func() (fs.FileInfo, error)
}

func (f File) Stat() (fs.FileInfo, error) {
	return f.stat()
}

type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (fi FileInfo) Name() string {
	return path.Base(fi.name)
}

func (fi FileInfo) Size() int64 {
	return fi.size
}

func (fi FileInfo) Mode() fs.FileMode {
	return fi.mode
}

func (fi FileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi FileInfo) IsDir() bool {
	return fi.mode.IsDir()
}

func (fi FileInfo) Sys() interface{} {
	return nil
}
