Bloque - A simple blocking queue
---

Bloque is a simple and straightforward concurrent blocking queue for Go. Most importantly, it provide context semantics to allow for blocked call timeout and cancellation. It makes the following guarantees:
* Strong concurrency: multiple go routines can safely access the queue.
* Blocking behavior: calling go routines will block if the operation cannot be satisfied.
* Context semantics: calls can safely be cancelled by their contexts.
* Strong ordering: blocked go routines will be unblocked in fifo manner.
* Graceful closing and draining: queue can be closed unblocking go routines until completely drained.

### Usage

```go
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Milliseconds)
    defer cancel()
    // this call will block until an item is available
    // or ctx is canceled.
    item, err := q.Pop(ctx)
```

It also provide multiple constrains to control growth:
|Constrain Option|Description|
|-|-|
|`WithCapacity`|Sets maximum queue capacity before `Push()` calls block.|
|`WithMaxPushWaiters`|Sets maximum goroutines blocked on `Push()`.|
|`WithMaxPopWaiters`|Sets maximum goroutines blocked on `Pop()`.|

```go
    // limit queue capacity to 10
    q := bloque.New(WithCapacity(10))
    // ...
    // multiple Push() and reached capacity 10
    //
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Milliseconds)
    defer cancel()
    // this call will block until an item is popped.
    err := q.Push(ctx, myItem)
```