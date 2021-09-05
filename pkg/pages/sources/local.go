package sources

import (
	"errors"
	"net/http"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/afero"
)

type LocalSource struct {
	BaseSource
	Root hcl.Expression `hcl:"root,optional"`
}

func (source *LocalSource) Fs(request *http.Request) (afero.Fs, error) {
	return source.GetCachedFs(request, func(context hcl.EvalContext) (afero.Fs, error) {
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

		return afero.NewBasePathFs(afero.NewOsFs(), rootDir), nil
	})
}

func (source *LocalSource) Configure() {
	source.GetFsCacheKeys().UseExpression(source.Root)
}

func init() {
	Register("local", &LocalSource{})
}
