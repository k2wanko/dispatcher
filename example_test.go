package dispatcher

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
)

func process(i int) QueueFunc {
	return func(ctx context.Context) error {
		select {
		case <-time.After(1 * time.Second):
			fmt.Printf("Count: %d\n", i)
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}
}

func ExampleCancel() {
	ctx, _ := context.WithTimeout(context.Background(), 50*time.Millisecond)
	d := New(ctx)
	for i := 0; i < 10; i++ {
		d.Add(process(i))
	}
	// Output:
	//
}
