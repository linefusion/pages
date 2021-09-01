package sources

import (
	"errors"
	"net/http"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/karlseguin/ccache/v2"
	"github.com/linefusion/pages/pkg/pages/cache"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/spf13/afero"
)

type LocalSource struct {
	cacheKeyFlags int
	cache         *ccache.Cache
	Root          hcl.Expression `hcl:"root,optional"`
}

func (source *LocalSource) Fs(r *http.Request) (afero.Fs, error) {
	key := cache.NewKeyBuilderFromRequest(r).UseFlags(source.cacheKeyFlags).GetString()

	var item *ccache.Item

	var fs afero.Fs
	item = source.cache.Get(key)
	if item == nil {
		context := config.CreateRequestContext(r)

		rootDir, err := os.Getwd()
		if err != nil {
			rootDir = "./"
		}

		if source.Root != nil {
			root, diagnostics := (source.Root).Value(&context)
			if diagnostics.HasErrors() {
				return nil, errors.New(diagnostics.Error())
			}

			if !root.IsNull() {
				rootDir = root.AsString()
			}
		}

		fs = afero.NewBasePathFs(afero.NewOsFs(), rootDir)
		source.cache.Set(key, fs, 0)
		return fs, nil
	}

	return item.Value().(afero.Fs), nil
}

func init() {
	Register("local", ParseLocalSource)
}

func ParseLocalSource(options hcl.Body) (Source, error) {
	context := config.CreateDefaultContext()

	var source LocalSource
	if diagnostics := gohcl.DecodeBody(options, &context, &source); diagnostics.HasErrors() {
		return nil, errors.New(diagnostics.Error())
	}

	keyBuilder := cache.NewKeyBuilder()
	if source.Root != nil {
		for _, variable := range (source.Root).Variables() {
			if variable.IsRelative() {
				continue
			}

			parts := []string{}
			for _, part := range variable {
				root, ok := part.(hcl.TraverseRoot)
				if ok {
					parts = append(parts, root.Name)
					continue
				}

				attr, ok := part.(hcl.TraverseAttr)
				if ok {
					parts = append(parts, attr.Name)
					continue
				}
			}

			if len(parts) < 2 || parts[0] != "request" {
				continue
			}

			switch parts[1] {
			case "method":
				keyBuilder.UseMethod()
			case "headers":
				keyBuilder.UseHeaders()
			case "scheme":
				keyBuilder.UseScheme()
			case "host":
				keyBuilder.UseHost()
			case "path":
				keyBuilder.UsePath()
			case "params":
				keyBuilder.UseParams()
			}
		}
	}

	source.cacheKeyFlags = keyBuilder.Flags()
	source.cache = ccache.New(ccache.Configure())

	return &source, nil
}
