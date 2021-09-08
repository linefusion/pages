package sources

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/linefusion/pages/pkg/pages/cache"
	"github.com/valyala/fasthttp"
	"github.com/zclconf/go-cty/cty"
)

type BaseSource struct {
	cacheKeys *cache.KeyBuilder
}

func (source *BaseSource) CacheKeys() *cache.KeyBuilder {
	if source.cacheKeys == nil {
		source.cacheKeys = cache.NewKeyBuilder()
	}
	return source.cacheKeys
}

func (source *BaseSource) CreateKey(request *fasthttp.Request) string {
	return source.CacheKeys().GetString(request)
}

func evaluate(context hcl.EvalContext, expr hcl.Expression, def cty.Value) (cty.Value, error) {
	if expr == nil {
		return def, nil
	}

	value, diagnostics := expr.Value(&context)
	if diagnostics.HasErrors() {
		return cty.NilVal, errors.New(diagnostics.Error())
	}

	return value, nil
}
