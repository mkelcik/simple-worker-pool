package main

import (
	"context"
	"fmt"
	"runtime"

	pool "github.com/mkelcik/simple-worker-pool"
)

type MyStruct struct {
	Value int
}

func main() {

	size := 10
	data := make([]MyStruct, size)
	for i := 0; i < size; i++ {
		data[i] = MyStruct{Value: i + 1}
	}

	for res := range pool.DoWorkFromSlice[MyStruct, string](context.Background(), data, func(ctx context.Context, value MyStruct) pool.JobResult[string] {
		return pool.NewResult[string](nil, fmt.Sprintf("%d", value.Value))
	}, uint(runtime.NumCPU())) {
		fmt.Println(res.Result())
	}
}
