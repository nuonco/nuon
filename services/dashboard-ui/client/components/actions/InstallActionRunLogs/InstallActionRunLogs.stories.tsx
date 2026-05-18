export default {
  title: 'Actions/InstallActionRunLogs',
}

import type { TActionConfig } from '@/types'
import { useLogFilters } from '@/hooks/use-log-filters'
import { LogStreamContext } from '@/providers/log-stream-provider'
import { InstallActionRunLogs } from './InstallActionRunLogs'

const noop = () => {}

const mockConfig: TActionConfig = {
  steps: [
    { id: 'step-1', name: 'build', idx: 0 },
    { id: 'step-2', name: 'deploy', idx: 1 },
    { id: 'step-3', name: 'verify', idx: 2 },
  ],
} as TActionConfig

const mockLogs = Array.from({ length: 5 }, (_, i) => ({
  id: `log-${i}`,
  body: `Log line ${i + 1}: running step...`,
  timestamp: new Date(Date.now() - i * 60000).toISOString(),
  severity_number: 9,
  severity_text: 'Info',
  service_name: 'runner',
  scope_name: 'oteljob',
  log_attributes: { workflow_step_name: 'build' },
})) as any

const mockLogStreamContext = {
  logs: mockLogs,
  logStreamId: 'log-stream-1',
  isLoading: false,
  error: null,
  connectionState: 'disconnected' as const,
}

const Providers = ({ children }: { children: React.ReactNode }) => (
  <LogStreamContext.Provider value={mockLogStreamContext}>
    {children}
  </LogStreamContext.Provider>
)

const VerticalStory = () => {
  const filters = useLogFilters(mockLogs)
  return (
    <Providers>
      <InstallActionRunLogs
        actionConfig={mockConfig}
        layout="vertical"
        allLogs={mockLogs}
        filteredLogs={filters.filteredLogs ?? []}
        isLoading={false}
        activeLog={undefined}
        handleActiveLog={noop}
        filters={filters}
      />
    </Providers>
  )
}
export { VerticalStory as Vertical }

const HorizontalStory = () => {
  const filters = useLogFilters(mockLogs)
  return (
    <Providers>
      <InstallActionRunLogs
        actionConfig={mockConfig}
        layout="horizontal"
        allLogs={mockLogs}
        filteredLogs={filters.filteredLogs ?? []}
        isLoading={false}
        activeLog={undefined}
        handleActiveLog={noop}
        filters={filters}
      />
    </Providers>
  )
}
export { HorizontalStory as Horizontal }

const failedStepLogs = [
  ...[0, 1, 2].map((i) => ({
    id: `fail-build-${i}`,
    body: `Compiling module ${i + 1}/3...`,
    timestamp: new Date(Date.now() - (10 - i) * 60000).toISOString(),
    severity_number: 9,
    severity_text: 'Info',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_attributes: { workflow_step_name: 'build' },
  })),
  ...[0, 1].map((i) => ({
    id: `fail-deploy-${i}`,
    body: i === 0 ? 'Starting deployment to cluster...' : 'Applying manifests...',
    timestamp: new Date(Date.now() - (7 - i) * 60000).toISOString(),
    severity_number: 9,
    severity_text: 'Info',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_attributes: { workflow_step_name: 'deploy' },
  })),
  {
    id: 'fail-deploy-warn',
    body: 'Warning: pod readiness check taking longer than expected',
    timestamp: new Date(Date.now() - 5 * 60000).toISOString(),
    severity_number: 13,
    severity_text: 'Warn',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_attributes: { workflow_step_name: 'deploy' },
  },
  {
    id: 'fail-deploy-err-1',
    body: 'Error: container "app" in pod "web-abc123" CrashLoopBackOff',
    timestamp: new Date(Date.now() - 4 * 60000).toISOString(),
    severity_number: 17,
    severity_text: 'Error',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_attributes: { workflow_step_name: 'deploy' },
  },
  {
    id: 'fail-deploy-err-2',
    body: 'Fatal: deployment rollout failed — deadline exceeded',
    timestamp: new Date(Date.now() - 3 * 60000).toISOString(),
    severity_number: 21,
    severity_text: 'Fatal',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_attributes: { workflow_step_name: 'deploy' },
  },
] as any

const failedStepConfig: TActionConfig = {
  steps: [
    { id: 'step-1', name: 'build', idx: 0 },
    { id: 'step-2', name: 'deploy', idx: 1 },
    { id: 'step-3', name: 'verify', idx: 2 },
    { id: 'step-4', name: 'notify', idx: 3 },
  ],
} as TActionConfig

const failedStepStatuses: Record<string, string> = {
  build: 'success',
  deploy: 'error',
}

const FailedProviders = ({ children }: { children: React.ReactNode }) => (
  <LogStreamContext.Provider value={{ ...mockLogStreamContext, logs: failedStepLogs }}>
    {children}
  </LogStreamContext.Provider>
)

const FailedStepStory = () => {
  const filters = useLogFilters(failedStepLogs)
  return (
    <FailedProviders>
      <InstallActionRunLogs
        actionConfig={failedStepConfig}
        layout="vertical"
        allLogs={failedStepLogs}
        filteredLogs={filters.filteredLogs ?? []}
        isLoading={false}
        activeLog={undefined}
        handleActiveLog={noop}
        filters={filters}
        stepStatuses={failedStepStatuses}
      />
    </FailedProviders>
  )
}
export { FailedStepStory as FailedStep }

const FailedStepHorizontalStory = () => {
  const filters = useLogFilters(failedStepLogs)
  return (
    <FailedProviders>
      <InstallActionRunLogs
        actionConfig={failedStepConfig}
        layout="horizontal"
        allLogs={failedStepLogs}
        filteredLogs={filters.filteredLogs ?? []}
        isLoading={false}
        activeLog={undefined}
        handleActiveLog={noop}
        filters={filters}
        stepStatuses={failedStepStatuses}
      />
    </FailedProviders>
  )
}
export { FailedStepHorizontalStory as FailedStepHorizontal }

const LoadingStory = () => {
  const filters = useLogFilters([])
  return (
    <Providers>
      <InstallActionRunLogs
        actionConfig={mockConfig}
        layout="vertical"
        filteredLogs={[]}
        isLoading={true}
        activeLog={undefined}
        handleActiveLog={noop}
        filters={filters}
      />
    </Providers>
  )
}
export { LoadingStory as Loading }
