package di

import (
	"fmt"
	"reflect"
	"sync"
)

type DIContainer struct {
	services     map[reflect.Type]reflect.Value
	constructors map[reflect.Type]reflect.Value
	mutex        sync.Mutex
}

func NewDIContainer() *DIContainer {
	return &DIContainer{
		services:     make(map[reflect.Type]reflect.Value),
		constructors: make(map[reflect.Type]reflect.Value),
		mutex:        sync.Mutex{},
	}
}

func RegisterService[T any](container *DIContainer, constructor func() T) {
	container.mutex.Lock()
	defer container.mutex.Unlock()

	serviceType := reflect.TypeOf((*T)(nil)).Elem()
	container.constructors[serviceType] = reflect.ValueOf(constructor)
}

func ResolveService[T any](container *DIContainer) (T, error) {
	var Zero T

	container.mutex.Lock()
	defer container.mutex.Unlock()

	if service, ok := container.services[reflect.TypeOf((*T)(nil)).Elem()]; ok == true {
		return service.Interface().(T), nil
	}

	constructor, ok := container.constructors[reflect.TypeOf(new(T)).Elem()]
	if ok == false {
		return Zero, fmt.Errorf("cannot resolve service of type %T", new(T))
	}

	constrType := constructor.Type()
	args := make([]reflect.Value, constrType.NumIn())
	for i := 0; i < constrType.NumIn(); i++ {
		argType := constrType.In(i)

		value, err := container.ResolveType(argType)
		if err != nil {
			return Zero, fmt.Errorf("cannot resolve type %v: %w", argType, err)
		}

		args[i] = value
	}

	res := constructor.Call(args)[0]
	container.services[reflect.TypeOf((*T)(nil)).Elem()] = res

	return res.Interface().(T), nil
}

func (c *DIContainer) ResolveType(sType reflect.Type) (reflect.Value, error) {
	var Zero reflect.Value

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if service, ok := c.services[sType]; ok == true {
		return service, nil
	}

	constructor, ok := c.constructors[sType]
	if ok == false {
		return Zero, fmt.Errorf("cannot resolve service of type %T", sType)
	}

	constrType := constructor.Type()
	args := make([]reflect.Value, constrType.NumIn())
	for i := 0; i < constrType.NumIn(); i++ {
		argType := constrType.In(i)

		value, err := c.ResolveType(argType)
		if err != nil {
			return Zero, fmt.Errorf("cannot resolve type %v: %w", argType, err)
		}

		args[i] = value
	}

	res := constructor.Call(args)[0]
	c.services[sType] = res

	return res, nil
}
