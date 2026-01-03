# Metrics Export

## Summary

Expose system metrics in OpenMetrics/Prometheus format for integration with
existing monitoring infrastructure. Enables hybrid observability: Pulse for
AI-powered insight, Prometheus for alerting and dashboards.

## Why

Fazt has Pulse for "cognitive observability" - natural language queries about
system state. But many users have existing monitoring stacks (Grafana,
Prometheus, Datadog). Rather than replace them, integrate with them.

**Philosophy alignment:**
- Single binary: Metrics export is ~100 lines, no deps
- JSON everywhere: `/metrics` is text, but `fazt.kernel.metrics()` returns JSON
- Events as spine: Metrics are derived from event stream

## Endpoint

```
GET /metrics
```

Returns OpenMetrics format (Prometheus-compatible):

```
# HELP fazt_requests_total Total HTTP requests
# TYPE fazt_requests_total counter
fazt_requests_total{app="blog",status="200"} 12453
fazt_requests_total{app="blog",status="404"} 23
fazt_requests_total{app="api",status="200"} 8901

# HELP fazt_request_duration_seconds Request latency
# TYPE fazt_request_duration_seconds histogram
fazt_request_duration_seconds_bucket{app="blog",le="0.01"} 10234
fazt_request_duration_seconds_bucket{app="blog",le="0.1"} 12100
fazt_request_duration_seconds_bucket{app="blog",le="1"} 12450
fazt_request_duration_seconds_bucket{app="blog",le="+Inf"} 12453
fazt_request_duration_seconds_sum{app="blog"} 234.56
fazt_request_duration_seconds_count{app="blog"} 12453

# HELP fazt_storage_bytes_total Database size in bytes
# TYPE fazt_storage_bytes_total gauge
fazt_storage_bytes_total 104857600

# HELP fazt_apps_total Number of deployed apps
# TYPE fazt_apps_total gauge
fazt_apps_total 12

# HELP fazt_events_total Events emitted
# TYPE fazt_events_total counter
fazt_events_total{type="http.request"} 21354
fazt_events_total{type="deploy.complete"} 47

# HELP fazt_uptime_seconds Seconds since start
# TYPE fazt_uptime_seconds counter
fazt_uptime_seconds 86423
```

## Metrics Exposed

### System

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `fazt_uptime_seconds` | counter | - | Time since start |
| `fazt_storage_bytes_total` | gauge | - | SQLite file size |
| `fazt_memory_bytes` | gauge | `type` | Memory usage (heap, stack) |
| `fazt_goroutines` | gauge | - | Active goroutines |

### Apps

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `fazt_apps_total` | gauge | - | Deployed app count |
| `fazt_requests_total` | counter | `app`, `status` | HTTP requests |
| `fazt_request_duration_seconds` | histogram | `app` | Latency distribution |
| `fazt_serverless_executions_total` | counter | `app` | JS handler runs |
| `fazt_serverless_errors_total` | counter | `app` | JS handler failures |

### Storage

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `fazt_kv_operations_total` | counter | `op` | KV get/set/delete |
| `fazt_ds_operations_total` | counter | `op` | Document store ops |
| `fazt_s3_bytes_total` | counter | `op` | Blob bytes read/written |

### Events

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `fazt_events_total` | counter | `type` | Events by type |
| `fazt_events_pending` | gauge | - | Unprocessed events |

### Network

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `fazt_proxy_requests_total` | counter | `domain` | Egress proxy requests |
| `fazt_proxy_bytes_total` | counter | `direction` | Bytes in/out |
| `fazt_tls_certs_total` | gauge | `status` | Certificates (valid/expiring) |

## JS API

```javascript
// Get metrics as JSON (for custom processing)
const metrics = await fazt.kernel.metrics();
// {
//   uptime_seconds: 86423,
//   storage_bytes: 104857600,
//   apps_total: 12,
//   requests: { blog: { "200": 12453, "404": 23 }, ... },
//   ...
// }

// Get specific metric
const uptime = await fazt.kernel.metric('uptime_seconds');
// 86423
```

## CLI

```bash
# Dump current metrics (text format)
fazt metrics

# Dump as JSON
fazt metrics --json

# Watch metrics (refresh every 5s)
fazt metrics --watch
```

## Prometheus Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'fazt'
    static_configs:
      - targets: ['fazt.example.com:443']
    scheme: https
    metrics_path: /metrics
```

## Authentication

The `/metrics` endpoint requires authentication by default:

```bash
# Enable public metrics (not recommended for production)
fazt config set metrics.public true

# Use bearer token
curl -H "Authorization: Bearer $TOKEN" https://fazt.example.com/metrics
```

## Implementation Notes

Uses Go's `expvar` pattern internally, formatted to OpenMetrics on output:

```go
type Metrics struct {
    mu sync.RWMutex
    counters map[string]*atomic.Int64
    gauges   map[string]*atomic.Int64
    histograms map[string]*Histogram
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    m.writeOpenMetrics(w)
}
```

**Binary impact:** ~100 lines, no external deps.

## Relationship to Pulse

| Concern | Pulse | Metrics |
|---------|-------|---------|
| Format | Natural language | Numeric time-series |
| Query | "How's the system?" | PromQL: `rate(fazt_requests_total[5m])` |
| Alerting | AI-generated insights | Threshold-based rules |
| Dashboard | None (CLI/API) | Grafana integration |

They complement each other:
- Metrics for automated alerting and dashboards
- Pulse for understanding *why* metrics changed
