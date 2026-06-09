import { useLogStreamData } from '@/hooks/use-logs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { CellTerminal } from './NotebookCellLogs'

interface INotebookCellLogs {
  logStreamId: string
  command?: string
  runCreatedAt?: string
  runUpdatedAt?: string
  isRunComplete?: boolean
  runFailed?: boolean
}

const CellTerminalContainer = ({
  command,
  runCreatedAt,
  runUpdatedAt,
  isRunComplete,
  runFailed,
}: Omit<INotebookCellLogs, 'logStreamId'>) => {
  const { logs, isLoading, connectionState } = useLogStreamData()

  return (
    <CellTerminal
      logs={logs}
      isLoading={isLoading}
      connectionState={connectionState}
      command={command}
      runCreatedAt={runCreatedAt}
      runUpdatedAt={runUpdatedAt}
      isRunComplete={isRunComplete}
      runFailed={runFailed}
    />
  )
}

export const NotebookCellLogsContainer = ({
  logStreamId,
  ...rest
}: INotebookCellLogs) => (
  <LogStreamProvider logStreamId={logStreamId}>
    <CellTerminalContainer {...rest} />
  </LogStreamProvider>
)
