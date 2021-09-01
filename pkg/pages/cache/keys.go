package cache

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"sort"
	"strconv"
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
	flags   int
	request *http.Request
}

func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{
		flags: 0,
	}
}

func NewKeyBuilderFromRequest(r *http.Request) *KeyBuilder {
	return &KeyBuilder{
		request: r,
		flags:   0,
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

// UseURL is an alias for .UseSchema().UseHost().UsePath()
func (key *KeyBuilder) UseURL() *KeyBuilder {
	return key.UseScheme().UseHost().UsePath()
}

func (key *KeyBuilder) UseAll() *KeyBuilder {
	return key.UseScheme().UseHost().UsePath().UseParams().UseHeaders()
}

func (key *KeyBuilder) GetUsingFlags(flags int) uint64 {
	hash := fnv.New64a()
	if key.request == nil {
		return 0
	}

	if flags&MethodFlag == MethodFlag {
		hash.Write([]byte(key.request.Method))
		hash.Write([]byte("\n"))
	}

	if flags&SchemeFlag == SchemeFlag {
		hash.Write([]byte(key.request.URL.Scheme))
		hash.Write([]byte("\n"))
	}

	if flags&HostFlag == HostFlag {
		hash.Write([]byte(key.request.URL.Host))
		hash.Write([]byte("\n"))
	}

	if flags&PathFlag == PathFlag {
		hash.Write([]byte(key.request.URL.Path))
		hash.Write([]byte("\n"))
	}

	if flags&ParamsFlag == ParamsFlag {
		params := key.request.URL.Query()
		fmt.Printf("     raw params: %s", key.request.URL.RawQuery)
		fmt.Printf("unsorted params: %s", params.Encode())
		for _, param := range params {
			sort.Slice(param, func(i, j int) bool {
				return param[i] < param[j]
			})
		}
		fmt.Printf("  sorted params: %s", params.Encode())
		hash.Write([]byte(params.Encode()))
	}

	if flags&HeadersFlag == HeadersFlag {
		keys := make([]string, 0, len(key.request.Header))
		for k := range key.request.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			hash.Write([]byte(k))
			hash.Write([]byte("\n"))
			for _, v := range key.request.Header[k] {
				hash.Write([]byte(v))
				hash.Write([]byte("\n"))
			}
		}
	}

	return hash.Sum64()
}

func (key *KeyBuilder) Get() uint64 {
	return key.GetUsingFlags(key.flags)
}

func (key *KeyBuilder) GetString() string {
	return strconv.FormatUint(key.Get(), 36)
}
