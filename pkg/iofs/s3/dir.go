package s3

import (
	"errors"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

var (
	_ fs.ReadDirFile = (*Dir)(nil)
)

type Dir struct {
	FileInfo
	s3cl   s3iface.S3API
	bucket string
	marker *string
	done   bool
	buf    []fs.DirEntry
	dirs   map[DirEntry]bool
}

func (dir *Dir) Stat() (fs.FileInfo, error) {
	return &dir.FileInfo, nil
}

func (dir *Dir) Read([]byte) (int, error) {
	return 0, &fs.PathError{
		Op:   "read",
		Path: dir.name,
		Err:  errors.New("is a directory"),
	}
}

func (dir *Dir) Close() error {
	return nil
}

func (dir *Dir) ReadDir(n int) (entries []fs.DirEntry, err error) {
	if n <= 0 {
		switch err := dir.readAll(); {
		case err == nil:
		case errors.Is(err, io.EOF):
			return []fs.DirEntry{}, nil
		default:
			return nil, err
		}

		entries, dir.buf = dir.buf, nil
		return entries, nil
	}

loop:
	for len(dir.buf) < n {
		switch err := dir.readNext(); {
		case err == nil:
			continue
		case errors.Is(err, io.EOF):
			break loop
		default:
			return nil, err
		}
	}

	offset := min(n, len(dir.buf))
	entries, dir.buf = dir.buf[:offset:offset], dir.buf[offset:]

	if dir.done && len(dir.buf) == 0 {
		err = io.EOF
	}

	return entries, err
}

func (dir *Dir) readAll() error {
	for !dir.done {
		switch err := dir.readNext(); {
		case err == nil:
			continue
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}
	}
	return io.EOF
}

func (dir *Dir) readNext() error {
	if dir.done {
		return io.EOF
	}

	name := strings.TrimRight(dir.name, "/")
	switch {
	case name == ".":
		name = ""
	default:
		name += "/"
	}

	out, err := dir.s3cl.ListObjects(&s3.ListObjectsInput{
		Bucket:    &dir.bucket,
		Delimiter: aws.String("/"),
		Prefix:    &name,
		Marker:    dir.marker,
	})
	if err != nil {
		return err
	}

	if dir.name != "." && len(out.CommonPrefixes)+len(out.Contents) == 0 {
		return &fs.PathError{
			Op:   "readdir",
			Path: strings.TrimSuffix(name, "/"),
			Err:  fs.ErrNotExist,
		}
	}

	dir.marker = out.NextMarker
	dir.done = out.IsTruncated != nil && !(*out.IsTruncated)

	if dir.dirs == nil {
		dir.dirs = make(map[DirEntry]bool)
	}

	for _, p := range out.CommonPrefixes {
		if p == nil || p.Prefix == nil {
			continue
		}

		de := DirEntry{
			FileInfo: FileInfo{
				name: path.Base(*p.Prefix),
				mode: fs.ModeDir,
			},
		}

		if _, ok := dir.dirs[de]; !ok {
			dir.dirs[de] = false
		}
	}

	for _, o := range out.Contents {
		if o == nil || o.Key == nil {
			continue
		}

		dir.buf = append(dir.buf, DirEntry{
			FileInfo: FileInfo{
				name:    path.Base(*o.Key),
				size:    derefInt64(o.Size),
				modTime: derefTime(o.LastModified),
			},
		})
	}

	dir.mergeDirFiles()

	if dir.done {
		return io.EOF
	}
	return nil
}

func (dir *Dir) mergeDirFiles() {
	if dir.buf == nil {
		// according to fs docs ReadDir should never return nil slice,
		// so we set it here.
		dir.buf = []fs.DirEntry{}
	}

	// we need a current len for sort.Search that doesn't change; otherwise
	// we could not append to the same slice.
	l := len(dir.buf)
	for de, used := range dir.dirs {
		if used {
			continue
		}

		i := sort.Search(l, func(i int) bool {
			return dir.buf[i].Name() >= de.Name()
		})

		if i == l && !dir.done {
			continue
		}
		dir.buf = append(dir.buf, de)
		dir.dirs[de] = true
	}

	sort.Slice(dir.buf, func(i, j int) bool {
		return dir.buf[i].Name() < dir.buf[j].Name()
	})
}

type DirEntry struct {
	FileInfo
}

func (entry DirEntry) Type() fs.FileMode {
	return entry.Mode().Type()
}

func (entry DirEntry) Info() (fs.FileInfo, error) {
	return entry.FileInfo, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func derefInt64(n *int64) int64 {
	if n != nil {
		return *n
	}
	return 0
}

func derefTime(t *time.Time) time.Time {
	if t != nil {
		return *t
	}
	return time.Time{}
}
