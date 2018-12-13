# gpool

goroutine pool

## Feature

* resize worker num
* worker idle timeout
* aysnc & sync mode

### Usage

```
package main

import (
	"fmt"
	"time"

	"github.com/rfyiamcool/gpool"
)

func main() {
	gp, err := gpool.NewGPool(&gpool.Options{
		MaxWorker:   5,
		MinWorker:   2,
		JobBuffer:   1,
		IdleTimeout: 120 * time.Second,
	})

	if err != nil {
		panic(err.Error())
	}

	for index := 0; index < 1000; index++ {
		gp.ProcessAsync(func() {
			fmt.Println(time.Now())
			time.Sleep(5 * time.Second)
		})
	}

	k := make(chan bool, 0)
	<- k
}

```
