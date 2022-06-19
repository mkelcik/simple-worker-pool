# Simple worker pool

Simple way to process your dataset via worker pool. You can use slice `DoWorkFromSlice` or channel `DoWork` as input to pool.
Callback then return `JobResult` or nil. If nil is returned nothing is written to result channel.

### How to use:
```go
// my dataset
var myDataSet []interface{}

poolSize := uint(10)

// can use channel pool.DoWork with channel or DoWorkFromSlice with slice
for result := range pool.DoWorkFromSlice(context.Background(), myDataSet, func(ctx context.Context, value interface{}) pool.JobResult {
    // do work here, then send job result
    // if nil is set as result, nothing is sent to result channel
    return pool.NewResult(nil, value)
}, poolSize) {
    // check for error in callback
    if result.Error() != nil {
		panic(result.Error())
    } 
	// then check my result result.Result()
}
```


## Examples:  
### Basic

```go
package main

import (
	"context"
	"fmt"
	"runtime"

	pool "github.com/mkelcik/simple-worker-pool"
)

func main() {
	// define number if workers in pool
	poolSize := uint(runtime.NumCPU())
	dataset := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}

	for result := range pool.DoWorkFromSlice(context.Background(), dataset, func(ctx context.Context, value interface{}) pool.JobResult {
		// correct type cast
		number, ok := value.(int)
		if !ok {
			return pool.NewResult(fmt.Errorf("wrong number `%v`, uint expected", value), nil)
		}
		return pool.NewResult(nil, number+number)
	}, poolSize) {
		// something bad happen in callback
		if result.Error() != nil {
			panic(result.Error())
		}
		fmt.Printf("%+v\n", result.Result())
	}
}

```

### Check Primes

Slice:
```go
package main

import (
	"context"
	"fmt"

	simpleWorkerPool "github.com/mkelcik/simple-worker-pool"
)

func main() {
	// define number if workers in pool
	poolSize := uint(5)

	// find primes in first 1000 numbers
	numbersToCheck := make([]interface{}, 1000)
	for k := range numbersToCheck {
		numbersToCheck[k] = uint(k)
	}

	primes := make([]uint, 0, len(numbersToCheck))

	// find primes via pool
	// if callback job is done, result is sent to result channel, and I just iterate over
	for result := range simpleWorkerPool.DoWorkFromSlice(context.Background(), numbersToCheck, func(ctx context.Context, value interface{}) simpleWorkerPool.JobResult {
		// correct type cast
		number, ok := value.(uint)
		if !ok {
			return simpleWorkerPool.NewResult(fmt.Errorf("wrong number `%v`, uint expected", value), nil)
		}

		// just simple dump check if number is prime (not optimised in any way, just need "something to do", check https://en.wikipedia.org/wiki/Prime_number)
		if number <= 1 {
			// if nil is sent as result, nothing is sent to result channel, it is just skipped
			return nil
		}

		for i := number - 1; i > 1; i-- {
			if number%i == 0 {
				return nil
			}
		}
		return simpleWorkerPool.NewResult(nil, number)
	}, poolSize) {
		// something bad happen in callback
		if result.Error() != nil {
			panic(result.Error())
		}

		// it is thread-save append slice here
		if prime, ok := result.Result().(uint); ok {
			primes = append(primes, prime)
		}
	}

	// primes are not sorted, because they are checked in parallel
	fmt.Printf("%+v", primes)
}
```

Channel:
```go
package main

import (
	"context"
	"fmt"

	pool "github.com/mkelcik/simple-worker-pool"
)

func main() {
	// define number if workers in pool
	poolSize := uint(5)

	// check first 1000 numbers
	sizeToCheck := uint(1000)

	primes := make([]uint, 0, sizeToCheck)

	// find primes via pool
	// if callback job is done, result is sent to result channel, and I just iterate over
	for result := range pool.DoWork(context.Background(), func() <-chan interface{} {
		// find primes in first 1000 numbers
		ch := make(chan interface{}, poolSize)
		go func() {
			defer close(ch)
			for i := uint(0); i <= uint(sizeToCheck); i++ {
				ch <- i
			}
		}()
		return ch
	}(), func(ctx context.Context, value interface{}) pool.JobResult {
		// correct type cast
		number, ok := value.(uint)
		if !ok {
			return pool.NewResult(fmt.Errorf("wrong number `%v`, uint expected", value), nil)
		}

		// just simple dump check if number is prime (not optimised in any way, just need "something to do", check https://en.wikipedia.org/wiki/Prime_number)
		if number <= 1 {
			// if nil is sent as result, nothing is sent to result channel, it is just skipped
			return nil
		}

		for i := number - 1; i > 1; i-- {
			if number%i == 0 {
				return nil
			}
		}
		return pool.NewResult(nil, number)
	}, poolSize) {
		// something bad happen in callback
		if result.Error() != nil {
			panic(result.Error())
		}

		// it is thread-save append slice here
		if prime, ok := result.Result().(uint); ok {
			primes = append(primes, prime)
		}
	}

	// primes are not sorted, because they are checked in parallel
	fmt.Printf("%+v", primes)
}

```
