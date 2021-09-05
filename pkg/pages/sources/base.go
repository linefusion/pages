package sources

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/hcl/v2"
	"github.com/karlseguin/ccache/v2"
	"github.com/linefusion/pages/pkg/pages/cache"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/spf13/afero"
)

type BaseSource struct {
	cacheKeys *cache.KeyBuilder
	cache     *ccache.Cache
}

func (base *BaseSource) GetFsCacheKeys() *cache.KeyBuilder {
	if base.cacheKeys == nil {
		base.cacheKeys = cache.NewKeyBuilder()
	}
	return base.cacheKeys
}

func (base *BaseSource) GetFsCache() *ccache.Cache {
	if base.cache == nil {
		base.cache = ccache.New(ccache.Configure())
	}
	return base.cache
}

func (source *BaseSource) GetCachedFs(request *http.Request, createFs func(hcl.EvalContext) (afero.Fs, error)) (afero.Fs, error) {
	cacheKey := source.GetFsCacheKeys().GetString(request)

	var fs afero.Fs
	var item *ccache.Item = source.GetFsCache().Get(cacheKey)

	if item != nil {
		return item.Value().(afero.Fs), nil
	}

	fmt.Println("Criando fs")

	context := config.CreateRequestContext(request)
	fs, err := createFs(context)
	if err != nil {
		return fs, err
	}

	source.cache.Set(cacheKey, fs, 0)
	return fs, nil
}
