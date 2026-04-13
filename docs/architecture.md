# GoScrapy Core Architecture

GoScrapy's data flow is designed for clarity and concurrent execution, utilizing a battle-tested architecture inspired by the Scrapy framework.

## Data Flow Diagram

```mermaid
flowchart LR
    %% Request Flow
    Spider -->|1. Request| Engine
    Engine -->|2. Schedule| Scheduler
    Scheduler -->|3. Pull Worker| WorkerQueue[(Worker Queue)]
    WorkerQueue -.->|4. Available Worker| Scheduler
    Scheduler -->|5. Pass Work| Worker
    Worker -->|6. Trigger| Executor
    Executor -->|7. Forward| Middlewares
    Middlewares -->|8. Request| HTTP_Client

    %% Response Flow
    HTTP_Client -.->|9. Response| Middlewares
    Middlewares -.->|10. Response| Executor
    Executor -.->|11. Callback| Spider

    %% Data Flow
    Spider ==>|12. Yield Record| Engine
    Engine ==>|13. Push Data| PipelineManager
    PipelineManager ==>|14. Export| Pipelines[(DB, CSV, File)]

    style Spider fill:#F5C4B3,stroke:#993C1D,stroke-width:1px,color:#711B0C
    style Engine fill:#B5D4F4,stroke:#185FA5,stroke-width:1px,color:#0C447C
    style Scheduler fill:#CECBF6,stroke:#534AB7,stroke-width:1px,color:#3C3489
    style WorkerQueue fill:#D3D1C7,stroke:#5F5E5A,stroke-width:1px,color:#444441
    style Worker fill:#9FE1CB,stroke:#0F6E56,stroke-width:1px,color:#085041
    style Executor fill:#FAC775,stroke:#854F0B,stroke-width:1px,color:#633806
    style Middlewares fill:#E5B8F3,stroke:#842B9E,stroke-width:1px,color:#4B1161
    style HTTP_Client fill:#C0DD97,stroke:#3B6D11,stroke-width:1px,color:#27500A
    style PipelineManager fill:#F4C0D1,stroke:#993556,stroke-width:1px,color:#72243E
    style Pipelines fill:#D3D1C7,stroke:#5F5E5A,stroke-width:1px,color:#444441
```

## Component Breakdown

1.  **Spider**: Your custom code that defines initial requests and parses responses.
2.  **Engine**: The central orchestrator that coordinates data flow between all components.
3.  **Scheduler**: Manages the request queue and handles worker distribution.
4.  **Worker**: Lightweight concurrent execution units that pull work from the scheduler.
5.  **Executor**: Manages the execution context of a request, including middleware triggering.
6.  **Middlewares**: Pluggable components that process requests/responses (e.g., retries, cookies).
7.  **Pipeline Manager**: Handles the post-processing and export of yielded items.
