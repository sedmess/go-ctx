package iter

type It[T any] interface {
	HasNext() bool
	Next() T
	Peek(func(T)) It[T]
}

func Map[T, R any](it It[T], mappingFunc func(val T) R) It[R] {
	return &itProxy[R]{
		hasNext: it.HasNext,
		next: func() R {
			val := it.Next()
			newVal := mappingFunc(val)
			return newVal
		},
	}
}

type ItMap[T, R any] interface {
	Map(func(T) R) It[R]
}

type ItSink[T any] interface {
	OnNext(T)
	Complete()
}

type itSinkProxy[T any] struct {
	onNext   func(val T)
	complete func()
}

type itProxy[T any] struct {
	hasNext func() bool
	next    func() T
}

func (it *itProxy[T]) HasNext() bool {
	return it.hasNext()
}

func (it *itProxy[T]) Next() T {
	return it.next()
}

func (it *itProxy[T]) Peek(onEach func(val T)) It[T] {
	return &itProxy[T]{
		hasNext: it.hasNext,
		next: func() T {
			val := it.next()
			onEach(val)
			return val
		},
	}
}

func (sink *itSinkProxy[T]) OnNext(t T) {
	sink.onNext(t)
}

func (sink *itSinkProxy[T]) Complete() {
	sink.complete()
}

func Single[T any](val T) It[T] {
	return Arr([]T{val})
}

func Arr[T any](slice []T) It[T] {
	i := 0
	return &itProxy[T]{
		hasNext: func() bool {
			return len(slice) > i
		},
		next: func() T {
			val := slice[i]
			i++
			return val
		},
	}
}

func Chan[T any](ch chan *T) It[T] {
	var nextVal *T
	received := false
	nextVal = nil
	receiveNext := func() {
		val := <-ch
		received = true
		nextVal = val
	}
	return &itProxy[T]{
		hasNext: func() bool {
			if received {
				return nextVal != nil
			}
			receiveNext()
			return nextVal != nil
		},
		next: func() T {
			if received {
				val := nextVal
				received = false
				nextVal = nil
				return *val
			}
			receiveNext()
			val := nextVal
			received = false
			nextVal = nil
			return *val
		},
	}
}

func Generate[T any](generator func() (T, bool)) It[T] {
	valCh := make(chan *T)
	go func() {
		for {
			val, last := generator()
			valCh <- &val
			if last {
				valCh <- nil
				break
			}
		}
	}()
	return Chan(valCh)
}

func Sink[T any](sinkConsumer func(sink ItSink[T])) It[T] {
	valCh := make(chan *T)
	sink := &itSinkProxy[T]{
		onNext: func(val T) {
			valCh <- &val
		},
		complete: func() {
			valCh <- nil
		},
	}
	go sinkConsumer(sink)
	return Chan(valCh)
}
