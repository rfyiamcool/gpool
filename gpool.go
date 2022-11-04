package gpool

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	DefaultGPoolMaxWorker = 5
	DefaultIdleTimeout    = 180 * time.Second
	DefaultDispatchPeriod = 200 * time.Millisecond

	bucketSize = 5
)

var (
	ErrJobChanFull  = errors.New("gpool job chan is full")
	ErrInvalidValue = errors.New("invalid value")

	GlobalGoPool *GoPool
)

type jobFunc func()

type taskModel struct {
	call jobFunc
	res  chan bool
}

type Options struct {
	// number of maxworkers,
	MaxWorker int
	MinWorker int
	// size of job queue
	JobBuffer      int
	IdleTimeout    time.Duration
	DispatchPeriod time.Duration
}

type GoPool struct {
	sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	isClose bool

	// opt *gpOptions
	maxWorker int
	curWorker int
	minWorker int

	idleTimeout    time.Duration
	dispatchPeriod time.Duration

	jobChan   chan taskModel
	jobBuffer int

	call jobFunc
}

func NewGPool(op *Options) (*GoPool, error) {
	pool := GoPool{}

	// cp option
	pool.maxWorker = op.MaxWorker
	pool.minWorker = op.MinWorker
	pool.jobBuffer = op.JobBuffer
	pool.idleTimeout = op.IdleTimeout
	pool.dispatchPeriod = op.DispatchPeriod

	// reset default var
	if pool.maxWorker < 5 {
		pool.maxWorker = DefaultGPoolMaxWorker
	}
	if pool.idleTimeout < time.Second {
		pool.idleTimeout = DefaultIdleTimeout
	}
	if pool.dispatchPeriod < time.Millisecond || pool.dispatchPeriod > time.Second {
		pool.dispatchPeriod = DefaultDispatchPeriod
	}
	if pool.minWorker <= 0 {
		pool.minWorker = 1
	}
	if pool.jobBuffer <= 0 {
		// jobBuffer must > 0, dispatch add goworker with the option
		pool.jobBuffer = pool.maxWorker
	}
	if pool.maxWorker < pool.minWorker {
		return &pool, errors.New("maxWorker must be more than minWorker")
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool.ctx = ctx
	pool.cancel = cancel
	pool.jobChan = make(chan taskModel, pool.jobBuffer)

	// start
	pool.spawnWorker(pool.maxWorker)
	return &pool, nil
}

// TODO: scan workers' timer in heap, kill worker
func (p *GoPool) dispatchRun(d time.Duration) {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-ticker.C:
			if len(p.jobChan) == 0 {
				continue
			}

			// don't need lock
			if p.maxWorker == p.curWorker {
				continue
			}

			// start some workers at a time
			p.spawnWorker(p.maxWorker / bucketSize)

		case <-p.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (p *GoPool) validator() error {
	if p.maxWorker <= 0 {
		return ErrInvalidValue
	}
	if p.idleTimeout.Seconds() <= 0 {
		return ErrInvalidValue
	}

	return nil
}

func (p *GoPool) spawnWorker(num int) {
	p.Lock()
	defer p.Unlock()

	if p.curWorker == p.maxWorker {
		return
	}

	for i := 0; i < num; i++ {
		if p.curWorker >= p.maxWorker {
			return
		}

		p.curWorker++
		go p.handler()
	}
}

func (p *GoPool) ResizeMaxWorker(maxWorker int) error {
	if maxWorker <= 0 || maxWorker < p.minWorker {
		return ErrInvalidValue
	}

	p.maxWorker = maxWorker
	return nil
}

func (p *GoPool) trySpawnWorker() {
	if len(p.jobChan) > 0 && p.maxWorker > p.curWorker {
		p.spawnWorker(p.maxWorker / bucketSize)
	}
}

// after push job queue, retrun direct
func (p *GoPool) ProcessAsync(f jobFunc) error {
	task := taskModel{
		call: f,
		res:  nil,
	}
	p.trySpawnWorker()

	select {
	case p.jobChan <- task:
		return nil
	}
}

// wait jobFunc finish
func (p *GoPool) ProcessSync(f jobFunc) error {
	task := taskModel{
		call: f,
		res:  make(chan bool, 1),
	}
	p.trySpawnWorker()

	select {
	case p.jobChan <- task:
		<-task.res
		return nil
	}
}

func (p *GoPool) handler() {
	timer := time.NewTimer(p.idleTimeout)
	defer timer.Stop()

	for {
		select {
		case job := <-p.jobChan:
			job.call()
			if job.res != nil {
				job.res <- true
			}

			timer.Reset(p.idleTimeout)

		case <-timer.C:
			// for a long time without a task, worker exit
			err := p.workerExit(timer)
			if err != nil {
				continue
			}

			return

		case <-p.ctx.Done():
			return
		}
	}
}

func (p *GoPool) workerExit(timer *time.Timer) error {
	p.Lock()
	defer p.Unlock()

	if p.curWorker <= p.minWorker {
		timer.Reset(p.idleTimeout)
		return errors.New("don't less than minWorker")
	}

	p.curWorker--
	timer.Stop()
	return nil
}

func (p *GoPool) Close() {
	p.Lock()
	defer p.Unlock()

	// double check
	if p.isClose {
		return
	}

	p.cancel()
	p.isClose = true
}
