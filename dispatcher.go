package dispatcher

import (
	"sync"

	"golang.org/x/net/context"
)

const (
	DefaultMaxWorker = 3
	DefaultMaxQueues = 10000
)

type (
	Dispatcher struct {
		HandleErrorFunc func(error)

		q      chan QueueFunc
		wg     sync.WaitGroup
		cacnel context.CancelFunc

		*options
	}

	QueueFunc func(context.Context) error

	options struct {
		maxWorker int
		maxQueues int
	}

	NewOption func(*options)
)

func (d *Dispatcher) MaxWorker() int {
	return d.maxWorker
}

func (d *Dispatcher) MaxQueues() int {
	return d.maxQueues
}

func (d *Dispatcher) QueueCount() int {
	return len(d.q)
}

func MaxWorker(max int) NewOption {
	return func(o *options) {
		o.maxWorker = max
	}
}

func MaxQueues(max int) NewOption {
	return func(o *options) {
		o.maxQueues = max
	}
}

func New(ctx context.Context, opts ...NewOption) *Dispatcher {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	opt := new(options)
	for _, o := range opts {
		o(opt)
	}

	if opt.maxWorker <= 0 {
		opt.maxWorker = DefaultMaxWorker
	}

	if opt.maxQueues <= 0 {
		opt.maxQueues = DefaultMaxQueues
	}

	d := &Dispatcher{
		q:       make(chan QueueFunc, opt.maxQueues),
		cacnel:  cancel,
		options: opt,
	}

	d.wg.Add(opt.maxWorker)
	for i := 0; i < opt.maxWorker; i++ {
		go func() {
			defer d.wg.Done()
			for {
				select {
				case f := <-d.q:
					if f == nil {
						return
					}
					d.sendErr(f(ctx))
				case <-ctx.Done():
					d.sendErr(ctx.Err())
					return
				}
			}
		}()
	}

	return d
}

func (d *Dispatcher) sendErr(err error) {
	if err == nil {
		return
	}

	f := d.HandleErrorFunc
	if f == nil {
		return
	}

	f(err)
}

func (d *Dispatcher) Add(f QueueFunc) {
	if f == nil {
		return
	}
	d.q <- f
}

func (d *Dispatcher) Close() {
	close(d.q)
	d.wg.Wait()
}
