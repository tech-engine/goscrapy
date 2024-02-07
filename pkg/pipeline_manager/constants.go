package pipelinemanager

// If true, all pipelines' Open method complete without an error,
// otherwise, pipeline manager won't start and return an error corresponding
// to the first pipeline to return an non-nil error.

// const PIPELINEMANAGER_PIPELINES_MUST_OPEN = false

// Reuseable Pipeline Item pool size
const PIPELINEMANAGER_ITEMPOOL_SIZE = 10000

// Max key-value pairs a Pipiline Item can have
const PIPELINEMANAGER_ITEM_SIZE = 24

// Output queue buffer size. Yield items are pushed to this queue,
// before being feed into the start of the pipelines.
const PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE = 0

// Max number of Outputs that will be allowed to processed concurrently in the pipeline
const PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY = 1000
