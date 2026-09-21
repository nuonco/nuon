export const DIFFS_WORKER_URL = '/assets/diffs-worker.js'

export const workerFactory = () =>
  new Worker(DIFFS_WORKER_URL, { type: 'module' })
