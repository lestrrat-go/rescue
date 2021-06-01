package rescue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/lestrrat-go/rescue"
)

func TestRescueDo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := rescue.New(ctx)

	go func(ctx context.Context) {
		defer rescue.Do(ctx)

		panic("foo")
	}(r.Context(ctx))

	if err := r.Error(ctx); err != nil {
		fmt.Println("pacni with", err)
	}
}

func TestRescueGo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := rescue.New(ctx)

	r.Go(ctx, func() {
		panic("foo")
	})

	if err := r.Error(ctx); err != nil {
		fmt.Println("pacni with", err)
	}
}

func TestRescueGroup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rg := rescue.NewGroup()

	var list []*rescue.Rescue
	for i := 0; i < 10; i++ {
		i := i

		r := rg.New(ctx)
		list = append(list, r)
		go func(ctx context.Context) {
			defer rescue.Do(ctx)
			if i%2 == 0 {
				panic("foo")
			}
		}(r.Context(ctx))
	}

	rg.Wait()

	for i, r := range list {
		if err := r.Error(ctx); err != nil {
			fmt.Printf("%d: panic with %s\n", i, err)
		}
	}
}
