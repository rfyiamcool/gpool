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
		MaxWorker:   5, 				// 最大的协程数
		MinWorker:   2, 				// 最小的协程数
		JobBuffer:   1, 				// 缓冲队列的大小
		IdleTimeout: 120 * time.Second, // 协程的空闲超时退出时间
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
