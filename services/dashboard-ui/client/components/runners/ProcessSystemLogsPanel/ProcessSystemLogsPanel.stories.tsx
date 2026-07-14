export default {
  title: 'Runners/ProcessSystemLogsPanel',
}

import { ProcessSystemLogsPanel } from './ProcessSystemLogsPanel'
import type { TRunnerProcess } from '@/types'

const baseProcess = {
  id: 'proc-123',
  type: 'runner',
  log_stream_id: 'log-1',
} as TRunnerProcess

export const Loading = () => <ProcessSystemLogsPanel isLoading runnerId="runner-1" />

export const NoLogStream = () => (
  <ProcessSystemLogsPanel
    process={{ ...baseProcess, log_stream_id: undefined } as TRunnerProcess}
    runnerId="runner-1"
  />
)
