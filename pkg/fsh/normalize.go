package fsh

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type SeparatorType int

const (
	SystemSeparator    SeparatorType = iota
	ForwardSeparator   SeparatorType = iota
	BackwardsSeparator SeparatorType = iota
)

type NormalizeOptions struct {
	Prefix           string
	Separator        SeparatorType
	TrimLeadingSlash bool
}

type NormalizeFS struct {
	Options NormalizeOptions
	Fs      fs.FS
}

func (fs *NormalizeFS) Normalize(parts ...string) string {
	if fs.Options.Prefix != "" {
		parts = append([]string{fs.Options.Prefix}, parts...)
	}

	separator := string(filepath.Separator)
	if fs.Options.Separator == ForwardSeparator {
		separator = "/"
	} else if fs.Options.Separator == BackwardsSeparator {
		separator = "\\"
	}

	p := filepath.Join(parts...)

	p = strings.ReplaceAll(p, string(filepath.Separator), separator)
	if fs.Options.TrimLeadingSlash {
		p = strings.TrimPrefix(p, separator)
	}

	return p
}

func (fs *NormalizeFS) Open(name string) (fs.File, error) {
	return fs.Fs.Open(fs.Normalize(name))
}

func Normalize(fs fs.FS, options NormalizeOptions) *NormalizeFS {
	return &NormalizeFS{
		Fs:      fs,
		Options: options,
	}
}
