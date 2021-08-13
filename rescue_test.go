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

	r := rescue.New()

	go func(ctx context.Context) {
		defer rescue.Do(ctx)

		panic("foo")
	}(r.Context(ctx))

	if err := r.Error(ctx); err != nil {
		fmt.Println("panic with", err)
	}
}

func TestRescueGo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := rescue.New()

	r.Go(ctx, func() {
		panic("foo")
	})

	if err := r.Error(ctx); err != nil {
		fmt.Println("panic with", err)
	}
}

func TestRescueGroup(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rg := rescue.NewGroup()

		for i := 0; i < 10; i++ {
			i := i

			r := rg.New()
			go func(ctx context.Context) (err error) {
				defer rescue.Do(rescue.Bind(ctx, &err))
				return fmt.Errorf("failure %d", i)
			}(r.Context(ctx))
		}

		rg.Wait()

		for err := range rg.Errors() {
			fmt.Printf("error with %s\n", err)
		}
	})
	t.Run("Panic", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rg := rescue.NewGroup()

		for i := 0; i < 10; i++ {
			i := i

			r := rg.New()
			go func(ctx context.Context) {
				defer rescue.Do(ctx)
				if i%2 == 0 {
					panic("foo")
				}
			}(r.Context(ctx))
		}

		rg.Wait()

		for err := range rg.Errors() {
			fmt.Printf("panic with %s\n", err)
		}
	})
}
