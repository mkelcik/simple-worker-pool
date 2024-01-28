# Simple worker pool

Simple way to process your dataset via worker pool. You can use slice `DoWorkFromSlice` or channel `DoWork` as input to pool.
Callback then return `JobResult` or nil. If nil is returned nothing is written to result channel.

### How to use:
```go
package main

import (
	"context"
	"fmt"
	"runtime"

	pool "github.com/mkelcik/simple-worker-pool/v2"
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
```


## Examples:  

See `examples` folder.
