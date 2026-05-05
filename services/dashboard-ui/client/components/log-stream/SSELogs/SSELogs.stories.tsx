export default {
  title: 'LogStream/SSELogs',
}

import { useCallback, useEffect, useRef, useState } from 'react'
import { SSELogs, LogsSkeleton } from './SSELogs'
import { LogPanel } from '@/components/log-stream/LogPanel'
import { UnifiedLogsContext } from '@/providers/unified-logs-provider'
import { LogStreamContext } from '@/providers/log-stream-provider'
import { useArrowKeys } from '@/hooks/use-arrow-keys'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TOTELLog, TLogStream } from '@/types'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'

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
    service_name: 'runner',
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
    service_name: 'runner',
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
    service_name: 'runner',
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
    service_name: 'runner',
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
      'runner_job.id': 'job3m9xkr7fvp2qtw4ycz1nbah',
      'runner_job.type': 'helm-deploy',
      'runner_job_execution.id': 'runhx5kcw8fvp3qtm2ycz9nbaj',
      'runner_job_execution_step.name': 'helm-upgrade',
      step: 'helm-upgrade',
      'helm.release': 'my-app',
      'helm.namespace': 'default',
      'helm.revision': '7',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
      'k8s.cluster.name': 'install-us-east-1',
      'k8s.namespace.name': 'default',
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
    body: 'Failed to connect to database: connection timeout after 30s — host=db-primary.us-east-1.rds.amazonaws.com port=5432 dbname=app_production sslmode=require',
    service_name: 'runner',
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
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'job3m9xkr7fvp2qtw4ycz1nbah',
      'runner_job.type': 'helm-deploy',
      'runner_job_execution.id': 'runhx5kcw8fvp3qtm2ycz9nbaj',
      'runner_job_execution_step.name': 'health-check',
      step: 'health-check',
      'error.type': 'ConnectionTimeout',
      'db.system': 'postgresql',
      'db.name': 'app_production',
      'net.peer.name': 'db-primary.us-east-1.rds.amazonaws.com',
      'net.peer.port': '5432',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
      'k8s.cluster.name': 'install-us-east-1',
      'k8s.namespace.name': 'default',
    },
    scope_attributes: {
      'otel.scope.name': 'oteljob',
      'otel.scope.version': 'v1',
    },
  },
  {
    id: 'log-6',
    timestamp: '2024-01-15T10:30:06.678Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:06.678',
    severity_number: 5,
    severity_text: 'Debug',
    body: 'Resolving terraform providers: hashicorp/aws v5.31.0, hashicorp/kubernetes v2.25.2',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobkz6m4npq8rvx1wts3ybaf2',
    runner_job_execution_id: 'runtz9m2npq5rvx7wkc4ybaj8',
    runner_job_execution_step: 'terraform-init',
    trace_id: 'e5f6a7b8c9d0e1f2a3b4c5d6a1b2c3d4',
    span_id: 'b5a6f7e8f1e2d3c4',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobkz6m4npq8rvx1wts3ybaf2',
      'runner_job.type': 'terraform-apply',
      'runner_job_execution.id': 'runtz9m2npq5rvx7wkc4ybaj8',
      'runner_job_execution_step.name': 'terraform-init',
      step: 'terraform-init',
      'terraform.provider.aws': 'v5.31.0',
      'terraform.provider.kubernetes': 'v2.25.2',
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
    id: 'log-7',
    timestamp: '2024-01-15T10:30:08.901Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:08.901',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Terraform plan: 3 to add, 1 to change, 0 to destroy',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobkz6m4npq8rvx1wts3ybaf2',
    runner_job_execution_id: 'runtz9m2npq5rvx7wkc4ybaj8',
    runner_job_execution_step: 'terraform-plan',
    trace_id: 'e5f6a7b8c9d0e1f2a3b4c5d6a1b2c3d4',
    span_id: 'a6f7e8f1e2d3c4b5',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobkz6m4npq8rvx1wts3ybaf2',
      'runner_job.type': 'terraform-apply',
      'runner_job_execution.id': 'runtz9m2npq5rvx7wkc4ybaj8',
      'runner_job_execution_step.name': 'terraform-plan',
      step: 'terraform-plan',
      'terraform.plan.add': '3',
      'terraform.plan.change': '1',
      'terraform.plan.destroy': '0',
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
    id: 'log-8',
    timestamp: '2024-01-15T10:30:10.234Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:10.234',
    severity_number: 21,
    severity_text: 'Fatal',
    body: 'Runner process crashed: out of memory — container limit 512Mi exceeded, peak usage 623Mi during terraform apply',
    service_name: 'runner',
    scope_name: 'system',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobkz6m4npq8rvx1wts3ybaf2',
    runner_job_execution_id: 'runtz9m2npq5rvx7wkc4ybaj8',
    runner_job_execution_step: 'terraform-apply',
    trace_id: 'e5f6a7b8c9d0e1f2a3b4c5d6a1b2c3d4',
    span_id: 'f7e8f1e2d3c4b5a6',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobkz6m4npq8rvx1wts3ybaf2',
      'runner_job.type': 'terraform-apply',
      'runner_job_execution.id': 'runtz9m2npq5rvx7wkc4ybaj8',
      'runner_job_execution_step.name': 'terraform-apply',
      step: 'terraform-apply',
      'error.type': 'OOMKilled',
      'container.memory.limit_bytes': '536870912',
      'container.memory.peak_bytes': '653262848',
    },
    resource_attributes: {
      'service.name': 'runner',
      'service.version': 'v0.14.82',
      'host.name': 'runner-7b8f4d6c9-xk2m4',
      'os.type': 'linux',
      'k8s.pod.name': 'runner-7b8f4d6c9-xk2m4',
      'k8s.container.name': 'runner',
      'k8s.namespace.name': 'nuon-runner',
    },
    scope_attributes: {
      'otel.scope.name': 'system',
      'otel.scope.version': 'v1',
    },
  },
  {
    id: 'log-9',
    timestamp: '2024-01-15T10:30:11.567Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:11.567',
    severity_number: 1,
    severity_text: 'Trace',
    body: 'HTTP GET /healthz 200 0.8ms',
    service_name: 'runner',
    scope_name: 'system',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'f6a7b8c9d0e1f2a3b4c5d6a1b2c3d4e5',
    span_id: 'e8f1e2d3c4b5a6f7',
    log_attributes: {
      'http.method': 'GET',
      'http.route': '/healthz',
      'http.status_code': '200',
      'http.duration_ms': '0.8',
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
    id: 'log-10',
    timestamp: '2024-01-15T10:30:13.890Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:13.890',
    severity_number: 13,
    severity_text: 'Warn',
    body: 'Kustomize build: deprecated field "patchesStrategicMerge" detected in kustomization.yaml, migrate to "patches" — see https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/patches/',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    runner_job_id: 'jobrx3m7npq4fvk8wtz1ycba2',
    runner_job_execution_id: 'runpx6m9npq2fvk5wtz8ycba3',
    runner_job_execution_step: 'kustomize-build',
    trace_id: 'a7b8c9d0e1f2a3b4c5d6a1b2c3d4e5f6',
    span_id: 'f1e2d3c4b5a6f7e8',
    log_attributes: {
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'jobrx3m7npq4fvk8wtz1ycba2',
      'runner_job.type': 'k8s-manifest',
      'runner_job_execution.id': 'runpx6m9npq2fvk5wtz8ycba3',
      'runner_job_execution_step.name': 'kustomize-build',
      step: 'kustomize-build',
      'kustomize.deprecated_field': 'patchesStrategicMerge',
      'kustomize.replacement_field': 'patches',
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
] as TOTELLog[]

const mockFilters: TLogFiltersProps = {
  selectedSeverities: new Set(['Trace', 'Debug', 'Info', 'Warn', 'Error', 'Fatal']),
  handleSeverityInputToggle: noop,
  handleSeverityButtonClick: noop,
  handleSeverityReset: noop,
  selectedServices: new Set(['runner']),
  availableServices: new Set(['runner']),
  handleServiceInputToggle: noop,
  handleServiceButtonClick: noop,
  handleServiceReset: noop,
  jobOutputOnly: false,
  handleJobOutputToggle: noop,
  searchQuery: '',
  sortDirection: 'desc',
  filteredLogs: mockLogs,
  handleSearchChange: noop,
  handleSortToggle: noop,
  handleSortChange: noop,
  filterStats: { selectedCount: 10, totalCount: 10 },
  sortStats: { direction: 'desc', isNewestFirst: true, isOldestFirst: false },
  severityStats: { selectedCount: 6, totalCount: 6 },
  serviceStats: { selectedCount: 1, totalCount: 1, isAllSelected: true },
} as unknown as TLogFiltersProps

const mockLogStream: TLogStream = {
  id: 'log-stream-1',
  org_id: 'org-mock-001',
  open: false,
} as TLogStream

const mockUnifiedContext = {
  logs: mockLogs,
  isLoading: false,
  error: null,
  connectionState: 'disconnected' as const,
  loadMore: noop,
  hasMore: false,
  isStreamOpen: false,
}

const Providers = ({ children }: { children: React.ReactNode }) => (
  <LogStreamContext.Provider value={{ logStream: mockLogStream, refresh: noop }}>
    <UnifiedLogsContext.Provider value={mockUnifiedContext}>
      {children}
    </UnifiedLogsContext.Provider>
  </LogStreamContext.Provider>
)

const useLogPanel = () => {
  const [activeLog, setActiveLog] = useState<TOTELLog | undefined>()
  const cycleDirectionRef = useRef<'up' | 'down' | undefined>()
  const { addPanel, updatePanel, removePanel } = useSurfaces()
  const panelIdRef = useRef<string | undefined>()

  const handleActiveLog = useCallback(
    (logId?: string) => {
      const log = logId ? mockLogs.find((l) => l.id === logId) : undefined
      setActiveLog(log)
    },
    []
  )

  useArrowKeys({
    onDownArrow() {
      if (!activeLog) return
      cycleDirectionRef.current = 'down'
      const idx = mockLogs.findIndex((l) => l.id === activeLog.id)
      const nextIdx = idx + 1 >= mockLogs.length ? 0 : idx + 1
      handleActiveLog(mockLogs[nextIdx]?.id)
    },
    onUpArrow() {
      if (!activeLog) return
      cycleDirectionRef.current = 'up'
      const idx = mockLogs.findIndex((l) => l.id === activeLog.id)
      handleActiveLog(mockLogs.at(idx - 1)?.id)
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
  const { activeLog, handleActiveLog } = useLogPanel()

  return (
    <Providers>
      <SSELogs
        filteredLogs={mockLogs}
        filters={mockFilters}
        activeLog={activeLog}
        handleActiveLog={handleActiveLog}
        loadMore={noop}
        hasMore={false}
        isLoading={false}
        isStreamOpen={false}
      />
    </Providers>
  )
}

export const WithLoadMore = () => {
  const { activeLog, handleActiveLog } = useLogPanel()

  return (
    <Providers>
      <SSELogs
        filteredLogs={mockLogs}
        filters={mockFilters}
        activeLog={activeLog}
        handleActiveLog={handleActiveLog}
        loadMore={noop}
        hasMore={true}
        isLoading={false}
        isStreamOpen={false}
      />
    </Providers>
  )
}

export const Loading = () => (
  <Providers>
    <SSELogs
      filteredLogs={[]}
      filters={mockFilters}
      activeLog={undefined}
      handleActiveLog={noop}
      loadMore={noop}
      hasMore={false}
      isLoading={true}
      isStreamOpen={false}
    />
  </Providers>
)

export const Skeleton = () => <LogsSkeleton />
