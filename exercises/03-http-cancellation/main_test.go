package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAnalyzeHandlerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/analyze", nil)
	recorder := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		defer close(done)
		analyzeHandler(recorder, req)
	}()

	time.Sleep(time.Second)
	cancel()

	select {
	case <-done:
		// Handler returned promptly after cancellation.
	case <-time.After(time.Second):
		t.Fatal("Handler did not return promptly after cancellation")
	}

	if recorder.Body.Len() != 0 {
		t.Fatalf("expected no success response after cancellation, got %q", recorder.Body.String())
	}
}
