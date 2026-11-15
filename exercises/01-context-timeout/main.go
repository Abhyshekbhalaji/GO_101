package main

import (
	"context"
	"fmt"
	"time"
)

func analyze(ctx context.Context) error {
	fmt.Println("Starting AI analysis...")
	select {
	case <-ctx.Done():
		fmt.Println("Analysis canceled:", ctx.Err())
		return ctx.Err()
	case <-time.After(5 * time.Second):
		fmt.Println("Analysis completed successfully")
		return nil
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := analyze(ctx); err != nil {
		fmt.Println("Error:", err)
	}
}
