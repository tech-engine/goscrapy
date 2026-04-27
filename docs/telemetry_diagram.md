# GoScrapy Telemetry Architecture

GoScrapy's telemetry system uses a decoupled, three-layer architecture to enable high-performance metric collection without blocking the crawler's hot-path.

## Architecture Overview

```mermaid
flowchart TD
    %% Collection Layer
    subgraph Collection ["Collection Layer (Non-Blocking)"]
        direction LR
        SM["Stats Middleware"]
        SCH["Scheduler"]
        WP["Worker Pool"]
        PM["Pipeline Manager"]
    end

    %% Aggregation Layer
    subgraph Aggregation ["Aggregation Layer (Periodic)"]
        Hub["Telemetry Hub"]
    end

    %% Exhibition Layer
    subgraph Exhibition ["Exhibition Layer (Broadcast)"]
        Obs["Observers (TUI/Logs)"]
    end

    %% Connections
    SM -.->|"Snapshot()"| Hub
    SCH -.->|"Snapshot()"| Hub
    WP -.->|"Snapshot()"| Hub
    PM -.->|"Snapshot()"| Hub
    Hub -->|"GlobalSnapshot"| Obs

    style Collection fill:#FFDFD3,stroke:#E27D60,stroke-width:2px
    style Aggregation fill:#B5D4F4,stroke:#185FA5,stroke-width:2px
    style Exhibition fill:#C0DD97,stroke:#3B6D11,stroke-width:2px
```

## Telemetry Flow Sequence

The following sequence highlights the decoupling between the high-speed recording path (async) and the periodic broadcasting path (ticker-based).

```mermaid
sequenceDiagram
    participant Components as Engine Components (Middleware/Scheduler/etc.)
    participant Hub as TelemetryHub
    participant Observer as Observer (TUI)

    Note over Components: Collection Phase (Non-Blocking)
    Components->>Components: Update internal atomic counters during execution
    
    Note over Hub, Observer: Broadcast Phase (Ticker: 500ms)
    loop Every Interval
        Hub->>Components: Snapshot()
        Components-->>Hub: ComponentSnapshot
        Hub->>Hub: Build GlobalSnapshot
        Hub->>Observer: OnSnapshot(GlobalSnapshot)
        Observer->>Observer: Render UI
    end
```

## Component Roles

| Component | Role | Implementation |
| :--- | :--- | :--- |
| **Engine Components** | Maintain their own local metrics without a central bottleneck. | `sync/atomic` counters inside `Scheduler`, `WorkerPool`, `PipelineManager`, etc. |
| **Stats Middleware** | Specifically captures HTTP request/response metrics (latency, status codes). | Pluggable middleware using atomic variables. |
| **TelemetryHub** | Orchestrates periodic polling of all registered components. | Background loop with a Go `time.Ticker`. |
| **IStatsObserver** | Consumes aggregated snapshots for visualization or export. | TUI (bubbletea) dashboard or standard logging. |
