# Task 08 — Add Prometheus metrics for worker pool saturation and goroutine count

## Goal

Expose operational visibility into the system's concurrency health — specifically whether the Forge is approaching saturation, how many requests are being dropped, and overall goroutine growth.

## Blocked by

- **#03** — Forge dispatcher must exist to instrument.

## Acceptance Criteria

- [ ] `github.com/prometheus/client_golang` is added to `go.mod`.
- [ ] `GET /metrics` is registered in `RegisterRoutes()` using `promhttp.Handler()`.
- [ ] The following metrics are emitted:

| Metric | Type | Description |
|--------|------|-------------|
| `forge_jobs_queued` | Gauge | Current number of jobs waiting in the channel |
| `forge_jobs_processed_total` | Counter | Total jobs successfully completed |
| `forge_jobs_dropped_total` | Counter | Total jobs rejected due to full queue (backpressure) |
| `http_requests_total` | Counter | Requests by `method`, `path`, `status` labels |
| `active_goroutines` | Gauge | Sampled from `runtime.NumGoroutine()` |

- [ ] `forge_jobs_queued` and `active_goroutines` are updated on a background ticker (every 5 seconds) rather than inline to avoid hot-path overhead.
- [ ] `forge_jobs_dropped_total` is incremented in `Dispatcher.Submit` on `ErrQueueFull`.
- [ ] A chi middleware records `http_requests_total` after each request.

## Implementation Notes

### Package structure

```
internal/metrics/
  metrics.go    — register and expose all vars
```

### Dependency

```
go get github.com/prometheus/client_golang
```

### Metrics registration

```go
var (
    ForgeJobsQueued = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "forge_jobs_queued",
        Help: "Current jobs waiting in the Forge channel.",
    })
    ForgeJobsProcessed = promauto.NewCounter(prometheus.CounterOpts{
        Name: "forge_jobs_processed_total",
        Help: "Total jobs completed by Forge workers.",
    })
    ForgeJobsDropped = promauto.NewCounter(prometheus.CounterOpts{
        Name: "forge_jobs_dropped_total",
        Help: "Total jobs rejected due to full queue.",
    })
    HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests.",
    }, []string{"method", "path", "status"})
    ActiveGoroutines = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_goroutines",
        Help: "Current goroutine count.",
    })
)
```

### Goroutine sampler

Start a goroutine in `NewServer()` that ticks every 5 seconds:

```go
go func() {
    t := time.NewTicker(5 * time.Second)
    for range t.C {
        metrics.ActiveGoroutines.Set(float64(runtime.NumGoroutine()))
    }
}()
```
