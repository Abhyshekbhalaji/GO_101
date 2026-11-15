# Go 101 — Production Go Learning Journal

Hands-on Go exercises and learning notes, growing toward production backend engineering for **TITAN INTELLIGENCE**.

> Learning status: Lesson 01 (Contexts, Goroutine Lifecycles & HTTP Cancellation) is complete at the conceptual level. The exercises below include a cancellable task, concurrent workers, and an HTTP handler cancellation test.

## Run the exercises

Requires **Go 1.22+**. No external dependencies.

```bash
git clone https://github.com/Abhyshekbhalaji/GO_101.git
cd GO_101
go run ./exercises/01-context-timeout
go run ./exercises/02-concurrent-workers
go test -v ./exercises/03-http-cancellation
```

## Lesson 01 — Context, goroutines and HTTP cancellation

### 1. What is `context.Context`?

A shared mechanism for propagating **cancellation, deadlines, and request-scoped values** through a call chain. A context is **not** a goroutine killer; operations must respond to the signal.

- `ctx.Done()` is a channel that closes when cancellation happens.
- `ctx.Err()` explains why: `context.Canceled` (manual/parent cancellation) or `context.DeadlineExceeded` (deadline fired first).
- `context.WithTimeout(parent, duration)` derives a context; use `defer cancel()` to release timer resources.
- The **first cancellation reason wins**: calling `cancel()` after a deadline doesn't change `ctx.Err()`.
- `context.Background()` is an uncancelable root; it does not automatically cancel network requests.
- `context.WithValue` is for request-scoped metadata, not application dependencies.

```text
parent context
     |
     +---- child context
     |         |
cancel parent  +---- child is canceled too
```

### 2. Exercise 01 — Five-second simulated task, two-second timeout

See [exercises/01-context-timeout/main.go](exercises/01-context-timeout/main.go).

**Original attempt:** `time.Sleep(5*time.Second)` before `select`. The sleep was a reasonable simulation of work, but it delayed cancellation observation until five seconds had elapsed.

**Correction:** Use `select` to wait for either `ctx.Done()` or a five-second timer. The function can then return after about two seconds.

**Important insight:** Returning from a function waiting on a result is not the same as canceling underlying work. A goroutine sleeping for five seconds will keep sleeping even if the caller stops waiting.

### 3. Exercise 02 — Three workers, shared deadline, WaitGroup

See [exercises/02-concurrent-workers/main.go](exercises/02-concurrent-workers/main.go).

Three goroutines simulate work taking 1, 3, and 5 seconds. All receive the same two-second context. Normally:

| Worker | Duration | Expected |
| --- | ---: | --- |
| 1 | 1s | Completes |
| 2 | 3s | Canceled |
| 3 | 5s | Canceled |

- `wg.Add(1)` registers work *before* launching a goroutine.
- `defer wg.Done()` indicates that the goroutine has returned.
- `wg.Wait()` waits for the counter to reach zero; it **does not cancel work**.
- With Go 1.22+ loop-variable semantics, closures capture a distinct per-iteration variable.

**Review:** 10/10 for the submitted implementation.

### 4. Exercise 03 — HTTP cancellation and httptest

See [exercises/03-http-cancellation/main.go](exercises/03-http-cancellation/main.go) and [main_test.go](exercises/03-http-cancellation/main_test.go).

The `GET /api/v1/analyze` handler models five seconds of work with a ten-second server timeout derived from `r.Context()`. The test attaches a cancellable context using `httptest.NewRequestWithContext`, invokes the handler concurrently, cancels after approximately one second, and verifies the handler exits promptly.

The handler's **initial submitted version** used a three-second server timeout and checked `ctx.Err()` after `select`; the reviewed version returns from the cancellation case directly. The test version uses a ten-second server timeout to isolate *client* cancellation.

```text
test goroutine                    handler goroutine
  |                                      |
  +-- start handler -------------------->|
  |                                      | select:
  | wait 1s                              | ctx.Done() OR task timer
  +-- cancel()                           |
  |                                      +-- observes cancellation; exits
  +-- receive on done channel <----------+-- close(done)
  |
  +-- PASS if return is prompt
```

**Submitted test outcome:** `PASS` (reported runtime about 1.00s), reviewed at 9.5/10. The sample includes a check that a successful JSON body was not returned on cancellation.

**What the test proves:** The handler returns promptly after its request context is canceled. It does *not* simulate an actual socket disconnection, nor prove that a remote AI provider stops processing.

### 5. Incoming vs outgoing HTTP requests

- A handler receives `(w http.ResponseWriter, r *http.Request)`. The incoming request's cancellation is exposed via `r.Context()`.
- `http.NewRequestWithContext(ctx, method, absoluteURL, body)` **constructs** an outgoing request.
- `http.Client.Do(req)` **sends** the outgoing request and normally respects its context. Reading the response body may also be interrupted.
- If the handler merely simulates work, it does **not** need to make another HTTP request to itself.
- A relative path such as `/api/v1/analyze` is not a valid URL for an ordinary outgoing `http.Client.Do` call.

### 6. When to use channels, contexts, WaitGroups, or mutexes

| Need | Primitive |
| --- | --- |
| Normal sequential result | Function return |
| Signal cancellation/deadline down a call chain | `context.Context` |
| Await all registered goroutines | `sync.WaitGroup` |
| Communicate results, jobs, or completion signals | Channel |
| Protect shared mutable data | `sync.Mutex` / `sync.RWMutex` |

In the test, `done := make(chan struct{})` is a **completion signal**, not a pointer into the handler. The handler goroutine explicitly closes it after returning. The test can `select` on that channel or a timeout.

### 7. Quiz and review

| Checkpoint | Answer | Outcome |
| --- | --- | --- |
| Timeout 2s vs simulated completion 5s: which `select` branch? | A — `ctx.Done()` | Correct |
| TITAN handler calling an external AI API: `r.Context()` + outgoing `Do(req)`? | A | Correct |
| Context final theoretical assessment | 5/6 | Passed after follow-up clarification |
| Timeout expires before later manual cancel: `ctx.Err()`? | C — `DeadlineExceeded` | Correct on final clarification |

**Learning review:** Strong practical understanding of context, `select`, WaitGroups, incoming HTTP handlers, outgoing HTTP clients, and test completion signals. Codex assisted parts of the cancellation test; independent understanding was discussed, and further unaided transfer exercises will help confirm mastery.

### 8. What remains

- [x] Cancellation mental model and five-second simulation
- [x] Concurrent workers with `sync.WaitGroup`
- [x] HTTP handler and cancellation test with `httptest`
- [ ] Add success-path and server-timeout tests
- [ ] Apply graceful shutdown to a real server
- [ ] Lesson 02 — Dependency Injection: handler → service → repository
- [ ] PostgreSQL access and transactions
- [ ] Worker pools, race detection, and production reliability

### Full lesson journal

[Notion — Go Production Lesson 01](https://app.notion.com/p/4190a39eefe181bbb4bbc943c5c82e39)

---
Exercises are educational examples, not a deployable AI service. The HTTP exercise's analysis is simulated with a timer and does not call a real model.