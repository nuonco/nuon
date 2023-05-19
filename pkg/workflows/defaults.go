package workflows

// each domain has it's own namespace, so we no longer need to split work by domain at the task queue level. By having a
// single queue in each namespace, we can more easily understand queue depth + have headroom to have a
// high-priority/low-priority queue in the future.
const DefaultTaskQueue string = "main"

// the api task queue is for api jobs
const APITaskQueue string = "api"

// each namespace has it's own queue for executors.
const ExecutorsTaskQueue string = "executors"
