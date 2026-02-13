package collections

import (
	"context"
	"fmt"

	"github.com/yeka-go/app"
)

type Instances[T any, U any] struct {
	instances    map[string]T
	configPrefix string
	builderFn    func(cfg U) (T, error)
	shutdownFn   func(obj T)
}

func (i *Instances[T, U]) Get(cmdContext context.Context, connectionName string) (T, error) {
	var emptyT T
	obj, ok := i.instances[connectionName]
	if ok {
		return obj, nil
	}

	configKey := i.configPrefix + "." + connectionName
	config := app.ConfigFromContext(cmdContext)

	var cfg U
	err := config.UnmarshalPath(configKey, cfg)
	if err != nil {
		return emptyT, fmt.Errorf("config.UnmarshalPath: %w", err)
	}

	obj, err = i.builderFn(cfg)
	if err != nil {
		return emptyT, err
	}

	i.instances[connectionName] = obj
	if i.shutdownFn != nil {
		i.shutdownFn(obj)
	}
	return obj, nil
}

func NewInstances[T any, U any](cfgPrefix string, builderFn func(cfg U) (T, error), shutdownFn func(obj T)) Instances[T, U] {
	return Instances[T, U]{
		instances:    map[string]T{},
		configPrefix: cfgPrefix,
		builderFn:    builderFn,
		shutdownFn:   shutdownFn,
	}
}

func SimpleBuilderWrapper[T any, U any](fn func(U) T) func(cfg U) (T, error) {
	return func(cfg U) (T, error) {
		return fn(cfg), nil
	}
}
