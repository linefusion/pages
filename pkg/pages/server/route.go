package server

import (
	"regexp"
	"strings"

	"github.com/karlseguin/ccache"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/linefusion/pages/pkg/pages/sources"
	"github.com/valyala/fasthttp"
)

type HostMatcher func(host string) map[string]string

type Page struct {
	cache  *ccache.Cache
	config config.PageBlock
	source sources.Source
	hosts  []HostMatcher
}

func NewRoute(cfg config.PageBlock) Page {
	if cfg.Path == "" {
		cfg.Path = "/"
	}

	route := Page{
		config: cfg,
		cache:  ccache.New(ccache.Configure()),
	}

	for _, host := range cfg.Hosts {
		if strings.HasPrefix(host, "/") && strings.HasSuffix(host, "/") {
			route.hosts = append(route.hosts, regexMatcher(host))
		} else {
			route.hosts = append(route.hosts, stringMatcher(host))
		}
	}

	source, err := sources.New(route.config.Source)
	if err != nil {
		panic(err)
	}

	route.source = source

	return route
}

func (route Page) IsFallback() bool {
	if len(route.config.Hosts) > 0 {
		return false
	}
	return route.config.Path == "/" || route.config.Path == ""
}

func (route Page) Matches(context *fasthttp.RequestCtx) (bool, map[string]string) {
	host := string(context.Host())

	if route.IsFallback() {
		return true, map[string]string{
			"_": host,
		}
	}

	for _, match := range route.hosts {
		matches := match(host)
		if len(matches) > 0 {
			return strings.HasPrefix(string(context.Path()), route.config.Path), matches
		}
	}

	return false, nil
}

func (route Page) Serve(context *fasthttp.RequestCtx, vars map[string]string) error {
	key := route.source.CreateKey(&context.Request)

	var handler Handler
	var item *ccache.Item = route.cache.Get(key)
	if item != nil {
		handler = item.Value().(Handler)
		return handler.Handle(context)
	}

	ctx := config.CreateRequestContext(&context.Request, vars)
	fs, err := route.source.CreateFs(&context.Request, ctx)
	if err != nil {
		return ErrCreateFs
	}

	handler = NewHandler(fs)
	route.cache.Set(key, handler, 0)

	return handler.Handle(context)
}

func stringMatcher(host string) HostMatcher {
	return func(h string) map[string]string {
		if h != host {
			return nil
		}
		return map[string]string{
			"_": h,
		}
	}
}

func regexMatcher(expr string) HostMatcher {
	regex := regexp.MustCompile(strings.TrimSuffix(strings.TrimPrefix(expr, "/"), "/"))
	names := regex.SubexpNames()
	return func(h string) map[string]string {
		matches := regex.FindStringSubmatch(h)
		if matches == nil {
			return nil
		}
		props := map[string]string{
			"_": matches[0],
		}
		for index, value := range matches {
			if names[index] == "" {
				continue
			}
			if index > 0 {
				props[names[index]] = value
			}
		}
		return props
	}
}
