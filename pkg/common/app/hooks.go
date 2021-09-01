package app

import "github.com/urfave/cli/v2"

type InitializerFunc func(*cli.Context) error
type FinalizerFunc func(*cli.Context) error

var (
	initializers []InitializerFunc = []InitializerFunc{}
	finalizers   []FinalizerFunc   = []FinalizerFunc{}
)

func GetInitializers() []InitializerFunc {
	return initializers
}

func AddInitializer(initializer InitializerFunc) {
	initializers = append(initializers, initializer)
}

func GetFinalizers() []FinalizerFunc {
	return finalizers
}

func AddFinalizer(finalizer FinalizerFunc) {
	finalizers = append(finalizers, finalizer)
}
