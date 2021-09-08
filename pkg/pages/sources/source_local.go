package sources

import (
	"errors"
	"io/fs"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/valyala/fasthttp"
)

type LocalSource struct {
	BaseSource
	Root hcl.Expression `hcl:"root,optional"`
}

func (source *LocalSource) CreateFs(request *fasthttp.Request, context hcl.EvalContext) (fs.FS, error) {
	rootDir, err := os.Getwd()
	if err != nil {
		rootDir = "./"
	}

	root, diagnostics := source.Root.Value(&context)
	if diagnostics.HasErrors() {
		return nil, errors.New(diagnostics.Error())
	}

	if !root.IsNull() {
		rootDir = root.AsString()
	}

	return os.DirFS(rootDir), nil
}

func (source *LocalSource) Configure() {
	source.CacheKeys().UseExpression(source.Root)
}

func init() {
	Register("local", &LocalSource{})
}
