package iofs

import "io/fs"

type FS interface {
	fs.FS
	fs.ReadDirFS
}
