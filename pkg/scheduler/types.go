package scheduler

type WorkQueue chan *schedulerWork
type WorkerQueue chan WorkQueue
