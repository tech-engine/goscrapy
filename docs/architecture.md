# GoScrapy Core Architecture

GoScrapy's data flow is designed for clarity and concurrent execution, utilizing a battle-tested architecture inspired by the Scrapy framework.

## Data Flow Diagram

```mermaid
flowchart LR
    Spider(((Spider)))
    
    subgraph Core [Engine Core]
        SignalBus{{Signal Bus}} -. Trigger .-> Engine{Engine}
        Engine -- 2. Schedule --> Scheduler[(Scheduler)]
        Scheduler -- 3. Next Request --> Engine
        Engine -- 4. Submit --> WorkerPool[Worker Pool]
        WorkerPool -- 9. Result --> Engine
    end

    Middlewares[[Middlewares]]
    HTTPAdapter([HTTP Adapter])
    
    subgraph Data [Data Flow]
        PipelineManager[Pipeline Manager] -- 13. Export --> Pipelines[(Pipelines)]
    end

    %% Flow
    Spider -- 1. Yield Requests --> Engine
    WorkerPool -- 5. Execute --> Middlewares
    Middlewares -- 6. Fetch --> HTTPAdapter
    HTTPAdapter -- 7. Response --> Middlewares
    Middlewares -- 8. Return --> WorkerPool
    Engine -- 10. Callback --> Spider
    Spider -- 11. Yield Items --> Engine
    Engine -- 12. Push --> PipelineManager

    %% Styling
    style SignalBus fill:#FFDFD3,stroke:#E27D60,stroke-width:2px,color:#8B4513
    style Spider fill:#F5C4B3,stroke:#993C1D,stroke-width:2px,color:#711B0C
    style Engine fill:#B5D4F4,stroke:#185FA5,stroke-width:2px,color:#0C447C
    style Scheduler fill:#CECBF6,stroke:#534AB7,stroke-width:1px,color:#3C3489
    style WorkerPool fill:#D3D1C7,stroke:#5F5E5A,stroke-width:1px,color:#444441
    style Middlewares fill:#E5B8F3,stroke:#842B9E,stroke-width:1px,color:#4B1161
    style HTTPAdapter fill:#C0DD97,stroke:#3B6D11,stroke-width:1px,color:#27500A
    style PipelineManager fill:#F4C0D1,stroke:#993556,stroke-width:1px,color:#72243E
    style Pipelines fill:#D3D1C7,stroke:#5F5E5A,stroke-width:1px,color:#444441
```

## Component Breakdown

1.  **Spider**: Your custom logic defining requests and parsing. Automatically discovered via reflection.
2.  **Engine**: The central orchestrator utilizing a **Signal Bus** for event-driven coordination.
3.  **Scheduler**: Manages the priority queue of pending requests.
4.  **Worker Pool**: A dynamic pool of workers that execute requests concurrently.
5.  **HTTP Adapter**: The network layer that performs the actual data fetching (e.g., standard HTTP, TLS spoofing).
6.  **Middlewares**: Pluggable hooks for modifying requests/responses (retries, cookies, stats).
7.  **Pipeline Manager**: Processes and exports items yielded by the spider.

## Signal-Driven Lifecycle

Starting with v0.26.0, GoScrapy uses a signal-based architecture to decouple core components. Instead of direct method calls between deep layers, the framework emits signals that interested components can subscribe to.

### Key Lifecycle Events:

- **SpiderOpened**: Triggered when the engine starts. This automatically calls the `Open(ctx)` method on your spider if it exists.
- **SpiderIdle**: Triggered when the engine detects no active requests or pending items. This is used for graceful shutdown.
- **SpiderClosed**: Triggered when the engine has finished all work.
- **ItemScraped/Dropped**: Triggered by the Pipeline Manager to notify observers of item progress.

This design allows for powerful extensions like the **TUI Dashboard** to monitor the crawler without modifying the core scraping logic.

## Auto-Discovery

GoScrapy minimizes boilerplate by automatically discovering and connecting your spider's methods to the signal bus. If your spider struct implements any of the following methods, they will be connected automatically:

- `Open(context.Context)`
- `Close(context.Context)`
- `Idle(context.Context)`
- `Error(context.Context, error)`

The engine uses reflection to map these methods to the internal signal bus during `RegisterSpider`, allowing you to focus on scraping logic rather than framework wiring.
