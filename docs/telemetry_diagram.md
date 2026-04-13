# GoScrapy Telemetry Architecture

GoScrapy's telemetry system uses a decoupled, three-layer architecture to enable high-performance metric collection without blocking the crawler's hot-path.

## Architecture Overview

```mermaid
flowchart TD
    %% Collection Layer
    subgraph Collection["Collection Layer (Concurrent)"]
        W1[Worker 1] -->|Records| R1[IStatsRecorder]
        W2[Worker 2] -->|Records| R2[IStatsRecorder]
        W3[Component] -->|Records| R3[IStatsRecorder]
    end

    %% Aggregation Layer
    subgraph Aggregation["Aggregation Layer (Periodic)"]
        R1 & R2 & R3 -.->|Injected via Context| Collector[IStatsCollector]
        Collector -->|Snapshot| Hub[TelemetryHub]
    end

    %% Exhibition Layer
    subgraph Exhibition["Exhibition Layer (Broadcast)"]
        Hub -->|GlobalSnapshot| Obs1[TUI Observer]
        Hub -->|GlobalSnapshot| Obs2[External Observer]
    end

    style Collection fill:#f9f,stroke:#333,stroke-width:2px
    style Aggregation fill:#bbf,stroke:#333,stroke-width:2px
    style Exhibition fill:#dfd,stroke:#333,stroke-width:2px
```

## Telemetry Flow Sequence

The following sequence highlights the decoupling between the high-speed recording path (async) and the periodic broadcasting path (ticker-based).

```mermaid
sequenceDiagram
    participant Worker as Crawler Worker
    participant Context as context.Context
    participant Collector as Aggregator (Collector)
    participant Hub as TelemetryHub
    participant Observer as IStatsObserver (TUI/Exposers)

    Note over Worker, Collector: Collection Phase (Hot-Path)
    Worker->>Context: FromContext(ctx)
    Context-->>Worker: IStatsRecorder
    Worker->>Worker: Execute Request
    Worker->>Collector: AddBytes(n) / AddSample(lat)
    
    Note over Hub, Observer: Broadcast Phase (Ticker: 500ms)
    loop Every Interval
        Hub->>Collector: Snapshot()
        Collector-->>Hub: ComponentSnapshot
        Hub->>Hub: Build GlobalSnapshot
        Hub->>Observer: OnSnapshot(GlobalSnapshot)
        Observer->>Observer: Update UI / Metrics
    end
```

## Component Roles

| Interface | Role | Lifecycle |
| :--- | :--- | :--- |
| **IStatsRecorder** | Captures individual events (bytes, duration). | Short-lived, per-request (via Context). |
| **IStatsCollector** | Aggregates recordings into a component-level state. | Long-lived, bound to Engine components. |
| **TelemetryHub** | Orchestrates periodic polling and broadcasting. | Singleton, bound to Engine lifecycle. |
| **IStatsObserver** | Consumes snapshots for visualization or export. | Pluggable (e.g., TUI, Prometheus). |
