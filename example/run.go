package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rfyiamcool/gpool"
)

var (
	wg = sync.WaitGroup{}
)

func main() {
	gp, err := gpool.NewGPool(&gpool.Options{
		MaxWorker:   5,                 // 最大的协程数
		MinWorker:   2,                 // 最小的协程数
		JobBuffer:   1,                 // 缓冲队列的大小
		IdleTimeout: 120 * time.Second, // 协程的空闲超时退出时间
	})

	if err != nil {
		panic(err.Error())
	}

	for index := 0; index < 1000; index++ {
		wg.Add(1)
		idx := index
		gp.ProcessAsync(func() {
			fmt.Println(idx, time.Now())
			time.Sleep(1 * time.Second)
			wg.Done()
		})
	}

	wg.Done()
}
