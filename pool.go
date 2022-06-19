package simple_worker_pool

import (
	"context"
	"sync"
)

type JobResult interface {
	Result() interface{}
	Error() error
}

type AsyncPanicError struct {
	err error
}

func (AsyncPanicError) Result() interface{} {
	return nil
}

func (err AsyncPanicError) Error() error {
	return err.err
}

// NewResult just general result response
func NewResult(err error, value interface{}) Result {
	return Result{
		value: value,
		err:   err,
	}
}

type Result struct {
	value interface{}
	err   error
}

func (r Result) Result() interface{} {
	return r.value
}

func (r Result) Error() error {
	return r.err
}

type CallbackFunc func(ctx context.Context, value interface{}) JobResult

// DoWorkFromSlice run async job from given slice
func DoWorkFromSlice(ctx context.Context, items []interface{}, callback CallbackFunc, workerSize uint) chan JobResult {

	// do not spawn more workers than size of items to process
	if uint(len(items)) < workerSize {
		workerSize = uint(len(items))
	}
	// create work chan
	work := make(chan interface{}, workerSize)

	// create worker pool
	res := DoWork(ctx, work, callback, workerSize)

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
func DoWork(ctx context.Context, workCh <-chan interface{}, callback CallbackFunc, workerSize uint) chan JobResult {
	// make result channel same size
	res := make(chan JobResult, workerSize)

	// start wait group
	w := sync.WaitGroup{}
	workerFunc := func(ctx context.Context, work <-chan interface{}) {
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
							res <- AsyncPanicError{err: err.(error)}
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
