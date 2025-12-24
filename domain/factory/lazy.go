package factory

import "sync"

func newLazyInitializer[T any]() *lazyInitializer[T] {
	return &lazyInitializer[T]{
		o: &sync.Once{},
	}
}

type lazyInitializer[T any] struct {
	o      *sync.Once
	object T
	err    error
}

func (li *lazyInitializer[T]) Get(creationFunc func() (T, error)) (T, error) {
	li.o.Do(func() {
		li.object, li.err = creationFunc()
	})

	return li.object, li.err
}
