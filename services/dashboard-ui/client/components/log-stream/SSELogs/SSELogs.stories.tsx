export default {
  title: 'LogStream/SSELogs',
}

import { useCallback, useEffect, useRef, useState } from 'react'
import { SSELogs, LogsSkeleton } from './SSELogs'
import { LogPanel } from '@/components/log-stream/LogPanel'
import { LogStreamContext } from '@/providers/log-stream-provider'
import { useArrowKeys } from '@/hooks/use-arrow-keys'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useLogFilters } from '@/hooks/use-log-filters'
import type { TOTELLog } from '@/types'

const noop = () => {}

const mockLogs: TOTELLog[] = [
  {
    id: 'log-1',
    timestamp: '2024-01-15T10:30:00.123Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:00.123',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Starting OCI sync for component helm-chart: pulling 766121324316.dkr.ecr.us-west-2.amazonaws.com/orgrok933tcyzji01s7us3aeo3/app98e2wpzdxwoey393edtqj45:bldq7fplr1up5atx5zpxotbabm',
    service_name: 'runner.actions.working',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobypfusmjzcer1unneh0jso67',
    runner_job_execution_id: 'runmwxrdg4jesn8jc1jhdjdyym',
    runner_job_execution_step: 'oci-sync',
    trace_id: 'a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6',
    span_id: 'f1e2d3c4b5a6f7e8',
    trace_flags: 1,
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobypfusmjzcer1unneh0jso67',
      'runner_job.type': 'oci-sync',
      'runner_job_execution.id': 'runmwxrdg4jesn8jc1jhdjdyym',
      'runner_job_execution_step.name': 'oci-sync',
      'nuon.tool': 'helm',
      step: 'oci-sync',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
      'process.runtime.name': 'go',
      'process.runtime.version': 'go1.25.0',
    },
    scope_attributes: {
      'otel.scope.name': 'oteljob',
      'otel.scope.version': 'v1',
    },
  },
  {
    id: 'log-2',
    timestamp: '2024-01-15T10:30:01.456Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:01.456',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Pulling image: 766121324316.dkr.ecr.us-west-2.amazonaws.com/orgrok933tcyzji01s7us3aeo3/app98e2wpzdxwoey393edtqj45:bldq7fplr1up5atx5zpxotbabm',
    service_name: 'runner.actions.dimngest-controller',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobypfusmjzcer1unneh0jso67',
    runner_job_execution_id: 'runmwxrdg4jesn8jc1jhdjdyym',
    runner_job_execution_step: 'oci-sync',
    trace_id: 'a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6',
    span_id: 'f1e2d3c4b5a6f7e8',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobypfusmjzcer1unneh0jso67',
      'runner_job.type': 'oci-sync',
      'runner_job_execution.id': 'runmwxrdg4jesn8jc1jhdjdyym',
      'runner_job_execution_step.name': 'oci-sync',
      'nuon.tool': 'helm',
      step: 'oci-sync',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
    },
    scope_attributes: {
      'otel.scope.name': 'oteljob',
      'otel.scope.version': 'v1',
    },
  },
  {
    id: 'log-3',
    timestamp: '2024-01-15T10:30:02.789Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:02.789',
    severity_number: 13,
    severity_text: 'Warn',
    body: 'ECR token refresh: token expires in 4m30s, threshold is 5m — refreshing early',
    service_name: 'runner.actions.pod',
    scope_name: 'system',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobypfusmjzcer1unneh0jso67',
    runner_job_execution_id: 'runmwxrdg4jesn8jc1jhdjdyym',
    runner_job_execution_step: 'initialize',
    trace_id: 'b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6a1',
    span_id: 'e2d3c4b5a6f7e8f1',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobypfusmjzcer1unneh0jso67',
      'runner_job.type': 'oci-sync',
      'runner_job_execution.id': 'runmwxrdg4jesn8jc1jhdjdyym',
      'runner_job_execution_step.name': 'initialize',
      step: 'initialize',
      'ecr.token_ttl_seconds': '270',
      'ecr.refresh_threshold_seconds': '300',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
    },
    scope_attributes: {
      'otel.scope.name': 'system',
      'otel.scope.version': 'v1',
    },
  },
  {
    id: 'log-4',
    timestamp: '2024-01-15T10:30:04.012Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:04.012',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Helm release "my-app" upgraded successfully in namespace "default" (revision 7)',
    service_name: 'runner.actions.ALTER',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'job3m9xkr7fvp2qtw4ycz1nbah',
    runner_job_execution_id: 'runhx5kcw8fvp3qtm2ycz9nbaj',
    runner_job_execution_step: 'helm-upgrade',
    trace_id: 'c3d4e5f6a7b8c9d0e1f2a3b4c5d6a1b2',
    span_id: 'd3c4b5a6f7e8f1e2',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'nuon.tool': 'helm',
      step: 'helm-upgrade',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
    },
    scope_attributes: {
      'otel.scope.name': 'oteljob',
      'otel.scope.version': 'v1',
    },
  },
  {
    id: 'log-5',
    timestamp: '2024-01-15T10:30:05.345Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:05.345',
    severity_number: 17,
    severity_text: 'Error',
    body: 'Failed to connect to database: connection timeout after 30s',
    service_name: 'runner.actions.GRANT',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'job3m9xkr7fvp2qtw4ycz1nbah',
    runner_job_execution_id: 'runhx5kcw8fvp3qtm2ycz9nbaj',
    runner_job_execution_step: 'health-check',
    trace_id: 'd4e5f6a7b8c9d0e1f2a3b4c5d6a1b2c3',
    span_id: 'c4b5a6f7e8f1e2d3',
    log_attributes: {
      'nuon.tool': 'helm',
      step: 'health-check',
    },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
] as TOTELLog[]

const mockLogStreamContext = {
  logs: mockLogs,
  logStreamId: 'lgsf8k2m4npq1rvx3wtz6yba7',
  isLoading: false,
  error: null,
  connectionState: 'connected' as const,
}

const Providers = ({ children }: { children: React.ReactNode }) => (
  <LogStreamContext.Provider value={mockLogStreamContext}>
    {children}
  </LogStreamContext.Provider>
)

const useLogPanel = (logs: TOTELLog[] | null) => {
  const [activeLog, setActiveLog] = useState<TOTELLog | undefined>()
  const cycleDirectionRef = useRef<'up' | 'down' | undefined>()
  const { addPanel, updatePanel, removePanel } = useSurfaces()
  const panelIdRef = useRef<string | undefined>()

  const handleActiveLog = useCallback(
    (logId?: string) => {
      const log = logId ? (logs ?? []).find((l) => l.id === logId) : undefined
      setActiveLog(log)
    },
    [logs]
  )

  useArrowKeys({
    onDownArrow() {
      if (!activeLog || !logs?.length) return
      cycleDirectionRef.current = 'down'
      const idx = logs.findIndex((l) => l.id === activeLog.id)
      const nextIdx = idx + 1 >= logs.length ? 0 : idx + 1
      handleActiveLog(logs[nextIdx]?.id)
    },
    onUpArrow() {
      if (!activeLog || !logs?.length) return
      cycleDirectionRef.current = 'up'
      const idx = logs.findIndex((l) => l.id === activeLog.id)
      handleActiveLog(logs.at(idx - 1)?.id)
    },
  })

  useEffect(() => {
    if (activeLog) {
      const panel = (
        <LogPanel
          log={activeLog}
          cycleDirection={cycleDirectionRef.current}
          onClose={() => handleActiveLog(undefined)}
        />
      )
      if (panelIdRef.current) {
        updatePanel(panelIdRef.current, panel)
      } else {
        cycleDirectionRef.current = undefined
        panelIdRef.current = 'log-panel'
        addPanel(panel, undefined, 'log-panel')
      }
    } else if (panelIdRef.current) {
      cycleDirectionRef.current = undefined
      removePanel(panelIdRef.current)
      panelIdRef.current = undefined
    }
  }, [activeLog])

  return { activeLog, handleActiveLog }
}

export const Default = () => {
  const filters = useLogFilters(mockLogs)
  const { activeLog, handleActiveLog } = useLogPanel(filters.filteredLogs)

  return (
    <Providers>
      <SSELogs
        filteredLogs={filters.filteredLogs ?? []}
        filters={filters}
        activeLog={activeLog}
        handleActiveLog={handleActiveLog}
        isLoading={false}
        isConnected={true}
      />
    </Providers>
  )
}

export const Loading = () => {
  const filters = useLogFilters([])

  return (
    <Providers>
      <SSELogs
        filteredLogs={[]}
        filters={filters}
        activeLog={undefined}
        handleActiveLog={noop}
        isLoading={true}
        isConnected={false}
      />
    </Providers>
  )
}

import { LogsPageSkeleton } from './SSELogs'

export const Skeleton = () => <LogsSkeleton />

export const PageSkeleton = () => <LogsPageSkeleton />
