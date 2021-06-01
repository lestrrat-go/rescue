package rescue

import (
	"context"
	"sync"
)

type identRescue struct{}

type Rescue struct {
	ch chan interface{}
	wg *sync.WaitGroup
}

func New(ctx context.Context) *Rescue {
	return &Rescue{
		ch: make(chan interface{}, 1),
	}
}

func (r *Rescue) Context(parent context.Context) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, identRescue{}, r)
}

func (r *Rescue) Done() {
	close(r.ch)
	if wg := r.wg; wg != nil {
		wg.Done()
	}
}

func Do(ctx context.Context) {
	var r *Rescue
	if v := ctx.Value(identRescue{}); v != nil {
		r = v.(*Rescue)
		defer r.Done()
	}

	if err := recover(); err != nil {
		if r != nil {
			select {
			case <-ctx.Done():
				return
			case r.ch <- err:
			}
			return
		}
		panic(err)
	}
}

func (r *Rescue) Error(ctx context.Context) interface{} {
	select {
	case <-ctx.Done():
		return nil
	case v := <-r.ch:
		return v
	}
}

type RescueGroup struct{
	wg *sync.WaitGroup
}

func NewGroup() *RescueGroup {
	return &RescueGroup{wg: &sync.WaitGroup{}}
}

func (rg *RescueGroup) New(ctx context.Context) *Rescue {
	r := New(ctx)
	r.wg = rg.wg
	rg.wg.Add(1)
	return r
}

func (rg *RescueGroup) Wait() {
	rg.wg.Wait()
}
