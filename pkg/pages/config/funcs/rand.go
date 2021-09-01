package funcs

import (
	"math/rand"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var RandFunc = function.New(
	&function.Spec{
		Params: []function.Parameter{},
		Type:   function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.NumberIntVal(rand.Int63()), nil
		},
	},
)
