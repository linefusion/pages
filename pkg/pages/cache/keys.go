package cache

import (
	"hash/fnv"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	"github.com/valyala/fasthttp"
)

const (
	MethodFlag  = 1 << iota
	SchemeFlag  = 1 << iota
	HostFlag    = 1 << iota
	PathFlag    = 1 << iota
	ParamsFlag  = 1 << iota
	HeadersFlag = 1 << iota
	BodyFlag    = 1 << iota
)

type KeyBuilder struct {
	flags int
}

func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{
		flags: 0,
	}
}

func (key *KeyBuilder) Flags() int {
	return key.flags
}

func (key *KeyBuilder) UseFlags(flags int) *KeyBuilder {
	key.flags = flags
	return key
}

func (key *KeyBuilder) UseMethod() *KeyBuilder {
	key.flags = key.flags | MethodFlag
	return key
}

func (key *KeyBuilder) UseScheme() *KeyBuilder {
	key.flags = key.flags | SchemeFlag
	return key
}

func (key *KeyBuilder) UseHost() *KeyBuilder {
	key.flags = key.flags | HostFlag
	return key
}

func (key *KeyBuilder) UsePath() *KeyBuilder {
	key.flags = key.flags | PathFlag
	return key
}

func (key *KeyBuilder) UseParams() *KeyBuilder {
	key.flags = key.flags | ParamsFlag
	return key
}

func (key *KeyBuilder) UseHeaders() *KeyBuilder {
	key.flags = key.flags | HeadersFlag
	return key
}

func (key *KeyBuilder) UseExpression(expression hcl.Expression) *KeyBuilder {
	if expression == nil {
		return key
	}

	for _, variable := range expression.Variables() {
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
			key.UseMethod()
		case "headers":
			key.UseHeaders()
		case "scheme":
			key.UseScheme()
		case "host":
			key.UseHost()
		case "path":
			key.UsePath()
		case "params":
			key.UseParams()
		}
	}

	return key
}

// UseURL is an alias for .UseSchema().UseHost().UsePath()
func (key *KeyBuilder) UseURL() *KeyBuilder {
	return key.UseScheme().UseHost().UsePath()
}

func (key *KeyBuilder) UseAll() *KeyBuilder {
	return key.UseScheme().UseHost().UsePath().UseParams().UseHeaders()
}

func (key *KeyBuilder) Get(request *fasthttp.Request) uint64 {
	hash := fnv.New64a()
	if request == nil {
		return 0
	}

	if key.flags&MethodFlag == MethodFlag {
		hash.Write(request.Header.Method())
		hash.Write([]byte("\n"))
	}

	if key.flags&SchemeFlag == SchemeFlag {
		hash.Write(request.URI().Scheme())
		hash.Write([]byte("\n"))
	}

	if key.flags&HostFlag == HostFlag {
		hash.Write(request.Host())
		hash.Write([]byte("\n"))
	}

	if key.flags&PathFlag == PathFlag {
		hash.Write(request.URI().Path())
		hash.Write([]byte("\n"))
	}

	if key.flags&ParamsFlag == ParamsFlag {
		/*
			params.VisitAll()
			for _, param := range params {
				sort.Slice(param, func(i, j int) bool {
					return param[i] < param[j]
				})
			}
		*/
		hash.Write(request.URI().QueryString())
	}

	if key.flags&HeadersFlag == HeadersFlag {
		hash.Write(request.Header.RawHeaders())
		hash.Write([]byte("\n"))
		/*
			keys := make([]string, 0, len(request.Header))
			for k := range request.Header {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				hash.Write([]byte(k))
				hash.Write([]byte("\n"))
				for _, v := range request.Header[k] {
					hash.Write([]byte(v))
					hash.Write([]byte("\n"))
				}
			}
		*/
	}

	return hash.Sum64()
}

func (key *KeyBuilder) GetString(request *fasthttp.Request) string {
	return strconv.FormatUint(key.Get(request), 36)
}
