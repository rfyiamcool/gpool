package gpool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func incrCounter(c *counter) {
	c.Lock()
	c.counter++
	c.Unlock()
}

type counter struct {
	sync.Mutex
	counter int
}

func TestBaseOption(t *testing.T) {
	opt := Options{
		MaxWorker: 10,
		MinWorker: 100,
	}

	_, err := NewGPool(&opt)
	if err == nil {
		t.Fatal(err)
	}
}

func TestBaseRun(t *testing.T) {
	opt := Options{
		MaxWorker:      10,
		MinWorker:      3,
		JobBuffer:      3,
		IdleTimeout:    10 * time.Second,
		DispatchPeriod: 100 * time.Millisecond,
	}

	pool, err := NewGPool(&opt)
	if err != nil {
		t.Fatal(err)
	}

	num := 100
	joinRun(t, num, pool, 0)
}

func joinRun(t *testing.T, num int, pool *GoPool, blockTime time.Duration) {
	incr := new(counter)
	res := make(chan bool, num)

	for index := 0; index < num; index++ {
		pool.ProcessAsync(
			func() {
				incrCounter(incr)
				res <- true
				if blockTime > 0 {
					time.Sleep(blockTime)
				}
			},
		)
	}

	lc := 0
	timer := time.NewTimer(15 * time.Second)
	for {
		select {
		case <-res:
			lc++
			if lc == num {
				return
			}

		case <-timer.C:
			if lc != num {
				t.Fatal("counter error")
			}
			return
		}
	}
}

// slow func
func TestDispatch(t *testing.T) {
	opt := Options{
		MaxWorker:      10,
		MinWorker:      1,
		JobBuffer:      1,
		IdleTimeout:    1 * time.Second,
		DispatchPeriod: 2 * time.Millisecond,
	}

	pool, err := NewGPool(&opt)
	if err != nil {
		t.Fatal(err)
	}

	// for ensure, double wait
	time.Sleep(opt.IdleTimeout * 2)
	if pool.curWorker != pool.minWorker {
		t.Fatal("worker timeout error")
	}

	if pool.maxWorker < pool.minWorker {
		t.Fatal("maxWorker > minWorker")
	}

	// notice: debug stdout
	// go func() {
	// 	for {
	// 		print("max")
	// 		print(pool.maxWorker)
	// 		print("cur")
	// 		print(pool.curWorker)
	// 		time.Sleep(50 * time.Millisecond)
	// 	}
	// }()

	num := 20
	starTS := time.Now()
	joinRun(t, num, pool, 1*time.Second)
	cost := time.Since(starTS).Seconds()
	if int(cost) > num/opt.MinWorker {
		t.Logf("join run cost: %v", cost)
		t.Fatal("dispatche don't add worker")
	}
}

func TestResize(t *testing.T) {
	// to do
}

func TestClosePool(t *testing.T) {
	// to do
}

func print(msg interface{}) {
	fmt.Println(msg)
}
