# Pulse - Cognitive Observability

## Summary

Pulse is Fazt's self-awareness system. It continuously monitors system health,
synthesizes metrics into understanding, and can explain what's happening in
natural language. Think of it as giving Fazt a "consciousness" about its own
state.

## Why Kernel-Level

Pulse is a fundamental OS primitive:
- Health monitoring is core to any operating system
- Must have access to all subsystems (proc, net, storage, fs)
- Runs continuously as a kernel service
- Other systems depend on it (alerts, fleet management)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         PULSE                                │
│                                                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ Collect │→│Synthesize│→│  Store  │→│   Act   │        │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │
│       ↑                                      │              │
│       │         Every 15 minutes             │              │
│       │                                      ▼              │
│  ┌─────────────────────────────────────────────────┐       │
│  │ Kernel Subsystems      │    Notifications       │       │
│  │ proc, net, fs,         │    (if critical)       │       │
│  │ storage, security      │                        │       │
│  └─────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────┘
```

## The Pulse Cycle

Every 15 minutes (configurable), Pulse executes a "beat":

### 1. Collect

Gather metrics from all kernel subsystems:

```go
type PulseMetrics struct {
    Timestamp   time.Time

    // System
    MemoryUsed  int64
    MemoryLimit int64
    CPUPercent  float64
    Uptime      time.Duration

    // Apps
    AppsActive  int
    AppsIdle    int
    AppsErroring int

    // Network
    RequestsTotal   int64
    RequestsErrored int64
    AvgLatencyMs    float64
    P99LatencyMs    float64

    // Storage
    StorageUsed  int64
    StorageLimit int64
    QueriesTotal int64

    // Per-App breakdown
    Apps []AppMetrics
}

type AppMetrics struct {
    UUID        string
    Name        string
    Requests    int64
    Errors      int64
    MemoryMB    int
    LastActive  time.Time
}
```

### 2. Synthesize (Cognitive)

If AI is configured (v0.12+), send metrics to LLM for analysis:

```go
type PulsePrompt struct {
    SystemContext  string      // Instance profile, limits, typical patterns
    CurrentMetrics PulseMetrics
    History        []PulseMetrics // Last 12 beats (3 hours)
    RecentEvents   []Event        // Deploys, errors, logins, etc.
}
```

**Prompt Template:**

```
You are the cognitive core of a Fazt personal cloud instance.

## Instance Profile
- Domain: {{.Domain}}
- Apps: {{.AppsActive}} active, {{.AppsIdle}} idle
- Owner timezone: {{.Timezone}}

## System Limits
- Memory: {{.MemoryLimit}} (using {{.MemoryUsed}})
- Storage: {{.StorageLimit}} (using {{.StorageUsed}})
- Max apps: {{.MaxApps}}

## Current State
- CPU: {{.CPUPercent}}%
- Memory: {{.MemoryUsed}}/{{.MemoryLimit}} ({{.MemoryPercent}}%)
- Requests (15m): {{.RequestsTotal}} ({{.RequestsErrored}} errors)
- Avg latency: {{.AvgLatencyMs}}ms (p99: {{.P99LatencyMs}}ms)

## Trend Data (15-minute buckets, last 3 hours)
| Time  | Requests | Errors | CPU  | Memory | Latency |
{{range .History}}
| {{.Time}} | {{.Requests}} | {{.Errors}} | {{.CPU}}% | {{.Memory}} | {{.Latency}}ms |
{{end}}

## Recent Events
{{range .Events}}
- {{.Time}}: {{.Description}}
{{end}}

## Task
Analyze this data and provide:
1. health: "healthy" | "degraded" | "critical"
2. issues: Array of {severity, title, description, recommendation}
3. patterns: Notable patterns or trends
4. security: Any security concerns
5. summary: One-sentence health summary

Respond as JSON matching the PulseAnalysis schema.
```

**Analysis Schema:**

```go
type PulseAnalysis struct {
    Health   string   `json:"health"` // healthy, degraded, critical
    Summary  string   `json:"summary"`
    Issues   []Issue  `json:"issues"`
    Patterns []string `json:"patterns"`
    Security []string `json:"security"`
}

type Issue struct {
    Severity       string `json:"severity"` // info, warning, critical
    Title          string `json:"title"`
    Description    string `json:"description"`
    Recommendation string `json:"recommendation"`
}
```

### 3. Store

Persist both raw metrics and analysis:

```sql
CREATE TABLE pulse_beats (
    id INTEGER PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    metrics_json TEXT NOT NULL,      -- Raw PulseMetrics
    analysis_json TEXT,               -- LLM analysis (nullable if no AI)
    health TEXT,                      -- healthy/degraded/critical
    created_at INTEGER DEFAULT (unixepoch())
);

CREATE INDEX idx_pulse_timestamp ON pulse_beats(timestamp);
CREATE INDEX idx_pulse_health ON pulse_beats(health);
```

Retention: Keep 7 days of beats (672 records at 15-min intervals).

### 4. Act

Based on analysis, take action:

```go
func (p *Pulse) Act(analysis PulseAnalysis) {
    // Critical issues → immediate notification
    for _, issue := range analysis.Issues {
        if issue.Severity == "critical" {
            p.alerts.Send(Alert{
                Severity: "critical",
                Title:    issue.Title,
                Message:  issue.Description,
                Action:   issue.Recommendation,
            })
        }
    }

    // Auto-mitigation (optional, configurable)
    if p.config.AutoMitigate {
        for _, security := range analysis.Security {
            if strings.Contains(security, "rate limit") {
                // Extract IP and apply temporary block
                p.net.RateLimit(extractIP(security))
            }
        }
    }
}
```

## Querying Pulse

### Natural Language Queries (Cognitive)

```javascript
// Ask Pulse about the system
const answer = await fazt.pulse.ask(
    "Why did errors spike around 2 PM?"
);

// Answer is LLM-generated from stored beats and analyses:
// "At 14:00, your checkout-app received a burst of traffic from a
// ProductHunt launch. The app wasn't warmed up, causing 503 errors
// for approximately 15 minutes. By 14:30, the app had scaled and
// errors returned to normal. Recommendation: Enable warmup handlers
// for traffic-sensitive apps."
```

Implementation:

```go
func (p *Pulse) Ask(question string) (string, error) {
    // Gather context
    recentBeats := p.getBeats(time.Now().Add(-24*time.Hour), time.Now())
    recentEvents := p.events.Query(24 * time.Hour)

    prompt := fmt.Sprintf(`
You are the cognitive core of a Fazt instance.

## Question
%s

## Recent Pulse Data (24 hours)
%s

## Recent Events
%s

Answer the question based on the data. Be specific and actionable.
If the data doesn't contain relevant information, say so.
`, question, formatBeats(recentBeats), formatEvents(recentEvents))

    return p.ai.Complete(prompt)
}
```

### Programmatic Queries

```javascript
// Get current status
const status = fazt.pulse.status();
// { health: "healthy", summary: "All systems nominal", lastBeat: "..." }

// Get recent history
const history = fazt.pulse.history(24); // Last 24 hours
// Returns array of PulseBeat objects

// Get insights (LLM-generated summaries)
const insights = fazt.pulse.insights(24);
// Returns array of notable insights from analyses

// Get specific metric trends
const memory = fazt.pulse.trend('memory', 24);
// Returns time-series data for charting
```

## Without AI (Basic Mode)

Pulse works without AI configured, but in reduced capacity:

| Feature | With AI | Without AI |
|---------|---------|------------|
| Metrics collection | Yes | Yes |
| Threshold alerts | Yes | Yes |
| Pattern detection | LLM-powered | Rule-based |
| Natural language queries | Yes | No |
| Anomaly detection | LLM-powered | Statistical |
| Recommendations | LLM-generated | Predefined |

Basic mode uses statistical anomaly detection:

```go
func (p *Pulse) detectAnomaliesBasic(current, history []PulseMetrics) []Issue {
    var issues []Issue

    // Calculate baselines
    avgRequests := average(history, "requests")
    avgErrors := average(history, "errors")
    avgLatency := average(history, "latency")

    // Check for deviations (>2 standard deviations)
    if current.Requests > avgRequests * 3 {
        issues = append(issues, Issue{
            Severity: "warning",
            Title: "Traffic spike",
            Description: fmt.Sprintf("Requests 3x above normal (%d vs %d avg)",
                current.Requests, avgRequests),
        })
    }

    // Memory threshold
    memPercent := float64(current.MemoryUsed) / float64(current.MemoryLimit)
    if memPercent > 0.9 {
        issues = append(issues, Issue{
            Severity: "critical",
            Title: "Memory exhaustion imminent",
            Description: fmt.Sprintf("Memory at %.0f%%", memPercent*100),
        })
    }

    return issues
}
```

## Configuration

```bash
# Set pulse frequency (default 15 minutes)
fazt config set pulse.interval 15m

# Enable/disable cognitive features
fazt config set pulse.cognitive true

# Set AI provider for analysis
fazt config set pulse.ai.provider anthropic
fazt config set pulse.ai.model claude-3-haiku  # Use fast model for frequent analysis

# Enable auto-mitigation
fazt config set pulse.auto_mitigate true

# Set alert thresholds
fazt config set pulse.thresholds.memory_critical 90
fazt config set pulse.thresholds.error_rate_warning 5
```

## JS API

```javascript
fazt.pulse.status()
// Returns: { health, summary, lastBeat, issues }

fazt.pulse.history(hours)
// Returns: PulseBeat[] - raw beats with metrics and analysis

fazt.pulse.insights(hours?)
// Returns: string[] - notable insights from LLM analyses

fazt.pulse.ask(question)
// Returns: string - LLM-generated answer about system state

fazt.pulse.trend(metric, hours)
// Returns: { timestamps: [], values: [] } - for charting

fazt.pulse.configure(options)
// Update pulse configuration
```

## CLI

```bash
# Current status
fazt pulse status
# Health: healthy
# Memory: 780MB / 1GB (78%)
# Active apps: 47
# Last 15m: 1,234 requests, 2 errors
# Summary: All systems nominal

# Ask a question
fazt pulse ask "What happened last night?"

# View history
fazt pulse history --hours 24

# View insights
fazt pulse insights

# Force a beat (manual trigger)
fazt pulse beat

# View trends
fazt pulse trend memory --hours 24 --format sparkline
# ▁▂▃▅▆▇█▇▆▅▃▂▁
```

## Admin SPA Integration

Pulse data powers the admin dashboard:

- **Health indicator**: Green/yellow/red based on `status.health`
- **Trend charts**: Memory, CPU, requests over time
- **Insights feed**: Recent LLM-generated insights
- **Ask box**: Natural language query input
- **Alert history**: Recent issues and resolutions

## Implementation Notes

- Pulse runs as a goroutine in the kernel
- Uses a ticker for beat intervals
- Metrics collected via existing kernel instrumentation
- AI calls use `fazt.ai` internally (v0.12 dependency for cognitive features)
- Beats stored in SQLite with automatic retention cleanup
- Ask queries rate-limited (1/minute default) to control AI costs

## Dependencies

- v0.8 Kernel: Metrics from proc, net, storage, fs
- v0.12 Agentic (optional): AI for cognitive features
- Alerts system (v0.8): For critical notifications

## Risks

- **AI Cost**: Frequent LLM calls add up. Mitigate with fast/cheap models.
- **False Positives**: LLM might over-interpret normal variations.
- **Storage**: 7 days of beats at 15-min intervals = ~672 rows. Manageable.
