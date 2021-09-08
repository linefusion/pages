package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"path/filepath"
	"strings"

	"github.com/valyala/fasthttp"
)

type Handler struct {
	fs         fs.FS
	indexPages []string
}

var (
	strSlashDotDotSlash = []byte("/../")
)

var (
	ErrFileExpected = errors.New("expecting a file, got something else")
	ErrInvalidPath  = errors.New("invalid path")
)

func NewHandler(fs fs.FS) Handler {
	handler := Handler{
		fs:         fs,
		indexPages: []string{"index.html", "index.htm"},
	}
	return handler
}

func (handler *Handler) Handle(context *fasthttp.RequestCtx) error {
	path := context.Request.URI().Path()
	pathStr := string(path)

	for _, index := range handler.indexPages {
		if strings.HasSuffix(pathStr, "/"+index) {
			qs := context.Request.URI().QueryString()
			newPath := []byte("./")
			if len(qs) > 0 {
				newPath = append(newPath, []byte("?")...)
				newPath = append(newPath, qs...)
			}
			context.Response.Header.SetCanonical([]byte("Location"), newPath)
			context.Response.SetStatusCode(fasthttp.StatusTemporaryRedirect) // Permanent
			return nil
		}
	}

	//hasTrailingSlash := len(path) > 0 && path[len(path)-1] == '/'
	path = stripTrailingSlashes(path)

	if n := bytes.IndexByte(path, 0); n >= 0 {
		return ErrInvalidPath
	}

	if n := bytes.Index(path, strSlashDotDotSlash); n >= 0 {
		return ErrInvalidPath
	}

	err := handler.serve(context, path)
	if err != nil {
		if err == ErrFileExpected {
			for _, index := range handler.indexPages {
				newPath := append(path, []byte(index)...)
				indexErr := handler.serve(context, newPath)
				if indexErr == nil {
					return nil
				}
			}
			return fs.ErrNotExist
		}
		return err
	}

	return nil
}

func (handler *Handler) serve(context *fasthttp.RequestCtx, path []byte) error {
	if len(path) == 0 {
		path = []byte(".")
	} else if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	f, err := handler.fs.Open(string(path))
	if err != nil {
		return err
	}

	stat, err := f.Stat()
	if err != nil {
		// TODO: extract and standardize
		return fmt.Errorf("cannot obtain info for file %q: %s", stat, err)
	}

	if stat.IsDir() {
		return ErrFileExpected
	}

	mimetype := mime.TypeByExtension(filepath.Ext(string(path)))
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}

	context.Response.SetStatusCode(200)
	context.Response.Header.SetContentLength(int(stat.Size()))
	context.Response.Header.SetContentType(mimetype)
	context.Response.Header.SetLastModified(stat.ModTime())
	context.Response.SetBodyStream(f, int(stat.Size()))
	return nil
}

func stripTrailingSlashes(path []byte) []byte {
	for len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}
