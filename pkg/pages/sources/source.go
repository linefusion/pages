package sources

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/hcl/v2"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/spf13/afero"
)

type Source interface {
	Fs(r *http.Request) (afero.Fs, error)
}

type SourceParser func(options hcl.Body) (Source, error)

var sourceParsers map[string]SourceParser = map[string]SourceParser{}

func Register(source string, parser SourceParser) {
	sourceParsers[source] = parser
}

func New(config config.PageSourceConfig) (Source, error) {
	parse, ok := sourceParsers[config.Type]
	if !ok {
		return nil, fmt.Errorf("unknown page source \"%s\"", config.Type)
	}
	return parse(config.Options)
}
