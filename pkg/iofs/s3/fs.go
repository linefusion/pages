package s3

import (
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/linefusion/pages/pkg/iofs"
)

var _ iofs.FS = (*FS)(nil)

var errNotDir = errors.New("not a dir")

// FS is a S3 filesystem implementation.
//
// S3 has a flat structure instead of a hierarchy. FS simulates directories
// by using prefixes and delims ("/"). Because directories are simulated, ModTime
// is always a default Time value (IsZero returns true).
type FS struct {
	client s3iface.S3API
	bucket string
	root   string
}

// New returns a new filesystem that works on the specified bucket.
func New(cl s3iface.S3API, bucket string, root string) *FS {
	root = strings.TrimSuffix(strings.TrimPrefix(root, "/"), "/")
	return &FS{
		client: cl,
		bucket: bucket,
		root:   root,
	}
}

// Open implements fs.FS.
func (f *FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	name = filepath.Join(f.root, name)
	name = strings.ReplaceAll(name, "\\", "/")

	if name == "." {
		return openDir(f.client, f.bucket, name)
	}

	out, err := f.client.GetObject(&s3.GetObjectInput{
		Key:    &name,
		Bucket: &f.bucket,
	})

	if err != nil {
		if isNotFoundErr(err) {
			switch d, err := openDir(f.client, f.bucket, name); {
			case err == nil:
				return d, nil
			case !isNotFoundErr(err) && !errors.Is(err, errNotDir) && !errors.Is(err, fs.ErrNotExist):
				return nil, err
			}

			return nil, &fs.PathError{
				Op:   "open",
				Path: name,
				Err:  fs.ErrNotExist,
			}
		}

		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  err,
		}
	}

	statFunc := func() (fs.FileInfo, error) {
		return stat(f.client, f.bucket, name)
	}

	if out.ContentLength != nil && out.LastModified != nil {
		// if we got all the information from GetObjectOutput
		// then we can cache fileinfo instead of making
		// another call in case Stat is called.
		statFunc = func() (fs.FileInfo, error) {
			return &FileInfo{
				name:    path.Base(name),
				size:    *out.ContentLength,
				modTime: *out.LastModified,
			}, nil
		}
	}

	return &File{
		ReadCloser: out.Body,
		stat:       statFunc,
	}, nil
}

// Stat implements fs.StatFS.
func (f *FS) Stat(name string) (fs.FileInfo, error) {
	fi, err := stat(f.client, f.bucket, name)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: name,
			Err:  err,
		}
	}
	return fi, nil
}

// ReadDir implements fs.ReadDirFS.
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	d, err := openDir(f.client, f.bucket, name)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: name,
			Err:  err,
		}
	}
	return d.ReadDir(-1)
}

func stat(s3cl s3iface.S3API, bucket, name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}

	if name == "." {
		return &Dir{
			s3cl:   s3cl,
			bucket: bucket,
			FileInfo: FileInfo{
				name: ".",
				mode: fs.ModeDir,
			},
		}, nil
	}

	out, err := s3cl.ListObjects(&s3.ListObjectsInput{
		Bucket:    &bucket,
		Delimiter: aws.String("/"),
		Prefix:    &name,
		MaxKeys:   aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}

	if len(out.CommonPrefixes) > 0 &&
		out.CommonPrefixes[0] != nil &&
		out.CommonPrefixes[0].Prefix != nil &&
		*out.CommonPrefixes[0].Prefix == name+"/" {
		return &Dir{
			s3cl:   s3cl,
			bucket: bucket,
			FileInfo: FileInfo{
				name: name,
				mode: fs.ModeDir,
			},
		}, nil
	}

	if len(out.Contents) != 0 &&
		out.Contents[0] != nil &&
		out.Contents[0].Key != nil &&
		*out.Contents[0].Key == name {
		return &FileInfo{
			name:    name,
			size:    derefInt64(out.Contents[0].Size),
			mode:    0,
			modTime: derefTime(out.Contents[0].LastModified),
		}, nil
	}

	return nil, fs.ErrNotExist
}

func openDir(s3cl s3iface.S3API, bucket, name string) (fs.ReadDirFile, error) {
	fi, err := stat(s3cl, bucket, name)
	if err != nil {
		return nil, err
	}

	if d, ok := fi.(fs.ReadDirFile); ok {
		return d, nil
	}
	return nil, errNotDir
}

var notFoundCodes = map[string]struct{}{
	s3.ErrCodeNoSuchKey: {},
	"NotFound":          {}, // localstack
}

func isNotFoundErr(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		_, ok := notFoundCodes[aerr.Code()]
		return ok
	}
	return false
}
