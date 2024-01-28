package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	simpleWorkerPool "github.com/mkelcik/simple-worker-pool/v2"
)

func main() {
	// define number if workers in pool
	poolSize := uint(runtime.NumCPU())

	// try to compare with processing only in one thread
	// poolSize := uint(1)

	// find primes_slice in first N numbers
	sizeToCheck := 100_000
	primes := make([]uint, 0, sizeToCheck)

	start := time.Now()
	// find primes via pool
	// if callback job is done, result is sent to result channel, and I just iterate over
	for result := range simpleWorkerPool.DoWork[uint, uint](context.Background(), func() <-chan uint {
		// find primes in first 1000 numbers
		ch := make(chan uint, poolSize)
		go func() {
			defer close(ch)
			for i := uint(1); i <= uint(sizeToCheck); i++ {
				ch <- i
			}
		}()
		return ch
	}(), func(ctx context.Context, number uint) simpleWorkerPool.JobResult[uint] {
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
		primes = append(primes, result.Result())
	}

	// primes_slice are not sorted, because they are checked in parallel
	fmt.Printf("%+v (done in %fs)", primes, time.Since(start).Seconds())
}
