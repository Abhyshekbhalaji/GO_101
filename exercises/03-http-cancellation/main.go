package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		fmt.Println("Analysis canceled:", ctx.Err())
		return
	case <-time.After(5 * time.Second):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"completed"}`))
	}
}

func main() {
	http.HandleFunc("/api/v1/analyze", analyzeHandler)
	fmt.Println("Server running on :8080")
	fmt.Println(http.ListenAndServe(":8080", nil))
}
