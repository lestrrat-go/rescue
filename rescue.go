package rescue

import (
	"context"
	"sync"
)

type identRescue struct{}

// Rescue act as a bridge between a groutine and its caller to communicate
// potential panic.
type Rescue struct {
	ch chan interface{}
	wg *sync.WaitGroup
}

// New creates a new Rescue instance
func New() *Rescue {
	return &Rescue{
		ch: make(chan interface{}, 1),
	}
}

// Context creates a new context.Context that can be passed to a goroutine.
// In order to propagate the panic to the caller, the context object must be
// passed to `rescue.Do` in a defer'ed statement so we can use `recover()`
func (r *Rescue) Context(parent context.Context) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, identRescue{}, r)
}

// Done marks the Rescue object as done. You cannot communicate over this
// object after `Done()` is called. Users usually do not need to call this
// as it is automatically called by `rescue.Do()`
func (r *Rescue) Done() {
	close(r.ch)
	if wg := r.wg; wg != nil {
		wg.Done()
	}
}

// Do detects panics caused in the current goroutine, and passes the value
// along to the `Rescue` object associated with the context object.
//
// Do must be called as `defer rescue.Do(ctx)`, as otherwise it cannot
// detect panics via `recover()`
//
// If a panic is encountered and the context object does not have a Rescue
// object associated with it, it will call `panic()` again.
func Do(ctx context.Context) {
	var r *Rescue
	if v := ctx.Value(identRescue{}); v != nil {
		r = v.(*Rescue)
		defer r.Done()
	}

	err := recover()
	if r != nil {
		select {
		case <-ctx.Done():
			return
		case r.ch <- err:
		}
		return
	}

	if err != nil {
		panic(err)
	}
}

// Go is a convenience function to wrap an anonymous function
// with rescue, so that you do not need to write `defer rescue.Do(ctx)` yourself.
//
// Generally it's probably better to change your goroutines to work with
// `Rescue` objects explicitly
func (r *Rescue) Go(ctx context.Context, fn func()) {
	go func(ctx context.Context) {
		defer Do(ctx)
		fn()
	}(r.Context(ctx))
}

// Error returns the value intercepted from a panic. If no panics
// are encountered, it returns nil
func (r *Rescue) Error(ctx context.Context) interface{} {
	select {
	case <-ctx.Done():
		return nil
	case v := <-r.ch:
		return v
	}
}

// RescueGroup is used to apply Rescue to multiple goroutines, and
// wait for their execution.
type RescueGroup struct{
	wg *sync.WaitGroup
}

// NewGroup creates a new RescueGroup
func NewGroup() *RescueGroup {
	return &RescueGroup{wg: &sync.WaitGroup{}}
}

// New creates a new Rescue that belongs to this RescueGroup
func (rg *RescueGroup) New() *Rescue {
	r := New()
	r.wg = rg.wg
	rg.wg.Add(1)
	return r
}

// Wait waits for all Rescue objects to be done
func (rg *RescueGroup) Wait() {
	rg.wg.Wait()
}
