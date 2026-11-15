package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func worker(ctx context.Context, id int, duration time.Duration) {
	fmt.Printf("Worker %d started\n", id)
	select {
	case <-ctx.Done():
		fmt.Printf("Worker %d canceled\n", id)
	case <-time.After(duration):
		fmt.Printf("Worker %d completed\n", id)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	durations := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}
	for i, duration := range durations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(ctx, i+1, duration)
		}()
	}
	wg.Wait()
	fmt.Println("All workers finished")
}
