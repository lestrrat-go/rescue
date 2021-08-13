# rescue

Aggregate errors/panics from goroutines

## Auto-capture panics (single goroutine)

```go
func Foo(ctx context.Context) {
  defer rescue.Do(ctx)
  ...
}

r := rescue.New()
go Foo(r.Context(ctx))


r.Wait()

if err := r.Error(ctx); err != nil {
  ... handle panic ...
}
```

## Auto-capture errors (single goroutine)

```go
func Foo(ctx context.Context) (err error){
  rescue.Bind(ctx, &err)
  defer rescue.Do(ctx)
  ...
}

r := rescue.New()
go Foo(r.Context(ctx))


r.Wait()

if err := r.Error(ctx); err != nil {
  // Because this shares the interface panics,
  // we need to convert types
  switch err := err.(type) {
  case error:
    ... handle error ...
  default:
    ... handle panic ...
  }
}
```

## Capture errors and panics within multiple goroutines

```go
func Foo(ctx context.Context) (err error){
  rescue.Bind(ctx, &err)
  defer rescue.Do(ctx)
  ...
}

rg := rescue.NewGroup()

for i := 0; i < 10; i++ {
  r := rg.New()
  go Foo(r.Context(ctx))
}

rg.Wait()

for err := rage rg.Errors() {
  // Because this shares the interface panics,
  // we need to convert types
  switch err := err.(type) {
  case error:
    ... handle error ...
  default:
    ... handle panic ...
  }
}

