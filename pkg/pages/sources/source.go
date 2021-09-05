package sources

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/karlseguin/ccache/v2"
	"github.com/linefusion/pages/pkg/pages/cache"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/spf13/afero"
)

type Source interface {
	Configure()

	Fs(request *http.Request) (afero.Fs, error)

	GetFsCache() *ccache.Cache
	GetFsCacheKeys() *cache.KeyBuilder
}

var registeredSources map[string]reflect.Type = map[string]reflect.Type{}

func Register(name string, source interface{}) {
	sourceType := reflect.TypeOf(source).Elem()

	if _, ok := source.(Source); !ok {
		panic("invalid source type")
	}
	registeredSources[name] = sourceType
}

func New(block config.SourceBlock) (Source, error) {
	sourceType, ok := registeredSources[block.Type]
	if !ok {
		return nil, fmt.Errorf("unknown page source \"%s\"", block.Type)
	}

	source := reflect.New(sourceType).Interface()
	err := configure(source, block.Options)
	if err != nil {
		return nil, err
	}

	src := source.(Source)
	src.Configure()

	return src, nil
}

func configure(source interface{}, options hcl.Body) error {
	context := config.CreateDefaultContext()
	if diagnostics := gohcl.DecodeBody(options, &context, source); diagnostics.HasErrors() {
		return errors.New(diagnostics.Error())
	}

	return nil
}
