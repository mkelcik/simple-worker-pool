package simple_worker_pool

import (
	"context"
	"sync"
)

type JobResult[T any] interface {
	Result() T
	Error() error
}

type AsyncPanicError[T any] struct {
	err   error
	value T
}

func (err AsyncPanicError[T]) Result() T {
	return err.value
}

func (err AsyncPanicError[T]) Error() error {
	return err.err
}

// NewResult just general result response
func NewResult[T any](err error, value T) Result[T] {
	return Result[T]{
		value: value,
		err:   err,
	}
}

type Result[T any] struct {
	value T
	err   error
}

func (r Result[T]) Result() T {
	return r.value
}

func (r Result[T]) Error() error {
	return r.err
}

type CallbackFunc[T any, K any] func(ctx context.Context, value T) JobResult[K]

// DoWorkFromSlice run async job from given slice
func DoWorkFromSlice[T, K any](ctx context.Context, items []T, callback CallbackFunc[T, K], workerSize uint) chan JobResult[K] {

	// do not spawn more workers than size of items to process
	if uint(len(items)) < workerSize {
		workerSize = uint(len(items))
	}
	// create work chan
	work := make(chan T, workerSize)

	// create worker pool
	res := DoWork[T](ctx, work, callback, workerSize)

	// send jobs to channel
	go func() {
		for _, v := range items {
			select {
			case <-ctx.Done():
				break
			case work <- v:
			}
		}
		close(work)
	}()

	// return result ch
	return res
}

// DoWork run async job from given channel
func DoWork[T, K any](ctx context.Context, workCh <-chan T, callback CallbackFunc[T, K], workerSize uint) chan JobResult[K] {
	// make result channel same size
	res := make(chan JobResult[K], workerSize)

	// start wait group
	w := sync.WaitGroup{}
	workerFunc := func(ctx context.Context, work <-chan T) {
		defer func() {
			w.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-work:
				// if all message is consumed and ch is closed ok is set to false
				if !ok {
					return
				}
				func() {
					defer func() {
						// recover from panic
						if err := recover(); err != nil {
							res <- AsyncPanicError[K]{err: err.(error)}
						}
					}()

					// if result is null, nothing is sent to result channel
					callBackResult := callback(ctx, value)
					if callBackResult == nil {
						return
					}

					select {
					// context done, exit
					case <-ctx.Done():
						return
					// wait for result channel is free and send
					case res <- callBackResult:
					}
				}()
			}
		}
	}

	// spawn workers
	for i := uint(0); i < workerSize; i++ {
		w.Add(1)
		go workerFunc(ctx, workCh)
	}

	// check job done
	go func() {
		w.Wait()
		close(res)
	}()
	return res
}
