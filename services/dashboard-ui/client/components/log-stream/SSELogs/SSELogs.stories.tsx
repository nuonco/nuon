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

const makeLog = (
  id: string,
  ts: string,
  severity: 'Trace' | 'Debug' | 'Info' | 'Warn' | 'Error' | 'Fatal',
  body: string,
  service = 'runner',
  scope = 'oteljob',
  tool = 'helm',
): TOTELLog =>
  ({
    id,
    timestamp: ts,
    timestamp_date: ts.slice(0, 10),
    timestamp_time: ts.slice(11, 23),
    severity_number: { Trace: 1, Debug: 5, Info: 9, Warn: 13, Error: 17, Fatal: 21 }[severity],
    severity_text: severity,
    body,
    service_name: service,
    scope_name: scope,
    log_stream_id: 'lgs-story',
    org_id: 'org-story',
    runner_id: 'rnr-story',
    runner_group_id: 'rgrp-story',
    runner_job_id: 'job-story',
    runner_job_execution_id: 'run-story',
    runner_job_execution_step: 'deploy',
    trace_id: 'aaaa',
    span_id: 'bbbb',
    log_attributes: { 'nuon.tool': tool },
    resource_attributes: { 'service.name': service },
    scope_attributes: { 'otel.scope.name': scope },
  }) as TOTELLog

const t = (base: string, offsetMs: number) => {
  const d = new Date(base)
  d.setMilliseconds(d.getMilliseconds() + offsetMs)
  return d.toISOString()
}

const BASE = '2026-05-26T14:00:00.000Z'

const helmDeployLogs: TOTELLog[] = [
  makeLog('h-001', t(BASE, 0), 'Info', 'Starting deploy for component "api-gateway"'),
  makeLog('h-002', t(BASE, 200), 'Info', 'Pulling OCI artifact 766121324316.dkr.ecr.us-west-2.amazonaws.com/org123/app456:bld789'),
  makeLog('h-003', t(BASE, 1200), 'Debug', 'OCI manifest digest: sha256:a3f8c9e2b1d04567890abcdef1234567890abcdef1234567890abcdef12345678'),
  makeLog('h-004', t(BASE, 1800), 'Info', 'Artifact pulled successfully (2.4 MB, 1.6s)'),
  makeLog('h-005', t(BASE, 2000), 'Info', 'Resolving helm chart dependencies'),
  makeLog('h-006', t(BASE, 2500), 'Debug', 'Dependency "postgresql" version 12.8.3 resolved from https://charts.bitnami.com/bitnami'),
  makeLog('h-007', t(BASE, 2800), 'Debug', 'Dependency "redis" version 17.15.2 resolved from https://charts.bitnami.com/bitnami'),
  makeLog('h-008', t(BASE, 3200), 'Info', 'All chart dependencies resolved (2 charts)'),
  makeLog('h-009', t(BASE, 3500), 'Info', 'Computing values diff against running release "api-gateway" (revision 12)'),
  makeLog('h-010', t(BASE, 3900), 'Debug', 'Values changed: image.tag "v2.3.0" → "v2.4.0", resources.limits.memory "512Mi" → "1Gi"'),
  makeLog('h-011', t(BASE, 4200), 'Info', 'Running pre-upgrade hooks'),
  makeLog('h-012', t(BASE, 4500), 'Debug', 'Hook "pre-upgrade/db-migrate" started (job api-gateway-migrate-xk2m4)'),
  makeLog('h-013', t(BASE, 5000), 'Info', 'Running database migration 20260520_add_audit_columns...'),
  makeLog('h-014', t(BASE, 6000), 'Info', 'Migration 20260520_add_audit_columns applied (1.0s)'),
  makeLog('h-015', t(BASE, 6500), 'Info', 'Running database migration 20260523_create_events_table...'),
  makeLog('h-016', t(BASE, 8200), 'Info', 'Migration 20260523_create_events_table applied (1.7s)'),
  makeLog('h-017', t(BASE, 8500), 'Info', 'Pre-upgrade hook "db-migrate" completed successfully'),
  makeLog('h-018', t(BASE, 9000), 'Info', 'Executing helm upgrade --install api-gateway ./chart --namespace production --wait --timeout 5m0s'),
  makeLog('h-019', t(BASE, 9500), 'Debug', 'Rendering templates with 47 value overrides'),
  makeLog('h-020', t(BASE, 10000), 'Debug', 'Template "deployment.yaml" rendered (2.1 KB)'),
  makeLog('h-021', t(BASE, 10200), 'Debug', 'Template "service.yaml" rendered (0.8 KB)'),
  makeLog('h-022', t(BASE, 10400), 'Debug', 'Template "hpa.yaml" rendered (0.6 KB)'),
  makeLog('h-023', t(BASE, 10600), 'Debug', 'Template "ingress.yaml" rendered (1.3 KB)'),
  makeLog('h-024', t(BASE, 10800), 'Debug', 'Template "configmap.yaml" rendered (3.7 KB)'),
  makeLog('h-025', t(BASE, 11000), 'Debug', 'Template "serviceaccount.yaml" rendered (0.4 KB)'),
  makeLog('h-026', t(BASE, 11200), 'Debug', 'Template "pdb.yaml" rendered (0.3 KB)'),
  makeLog('h-027', t(BASE, 11500), 'Info', '7 templates rendered, total manifest size 9.2 KB'),
  makeLog('h-028', t(BASE, 12000), 'Info', 'Applying manifests to cluster'),
  makeLog('h-029', t(BASE, 12500), 'Debug', 'ConfigMap "api-gateway-config" configured'),
  makeLog('h-030', t(BASE, 12800), 'Debug', 'ServiceAccount "api-gateway" unchanged'),
  makeLog('h-031', t(BASE, 13100), 'Debug', 'Service "api-gateway" unchanged'),
  makeLog('h-032', t(BASE, 13400), 'Debug', 'Deployment "api-gateway" configured'),
  makeLog('h-033', t(BASE, 13700), 'Debug', 'HorizontalPodAutoscaler "api-gateway" configured'),
  makeLog('h-034', t(BASE, 14000), 'Debug', 'Ingress "api-gateway" unchanged'),
  makeLog('h-035', t(BASE, 14300), 'Debug', 'PodDisruptionBudget "api-gateway" unchanged'),
  makeLog('h-036', t(BASE, 14500), 'Info', 'All manifests applied, waiting for rollout'),
  makeLog('h-037', t(BASE, 15000), 'Info', 'Waiting for deployment "api-gateway" rollout to finish: 0 of 4 updated replicas are available'),
  makeLog('h-038', t(BASE, 20000), 'Info', 'Waiting for deployment "api-gateway" rollout to finish: 1 of 4 updated replicas are available'),
  makeLog('h-039', t(BASE, 25000), 'Info', 'Waiting for deployment "api-gateway" rollout to finish: 2 of 4 updated replicas are available'),
  makeLog('h-040', t(BASE, 30000), 'Warn', 'Pod api-gateway-7b8f4d6c9-q3k9x: container "api-gateway" readiness probe failed: HTTP probe failed with statuscode: 503'),
  makeLog('h-041', t(BASE, 32000), 'Debug', 'Pod api-gateway-7b8f4d6c9-q3k9x: back-off restarting failed container (attempt 1)'),
  makeLog('h-042', t(BASE, 35000), 'Info', 'Pod api-gateway-7b8f4d6c9-q3k9x: container "api-gateway" readiness probe succeeded'),
  makeLog('h-043', t(BASE, 37000), 'Info', 'Waiting for deployment "api-gateway" rollout to finish: 3 of 4 updated replicas are available'),
  makeLog('h-044', t(BASE, 42000), 'Info', 'Deployment "api-gateway" successfully rolled out (4/4 replicas available)'),
  makeLog('h-045', t(BASE, 42500), 'Info', 'Running post-upgrade hooks'),
  makeLog('h-046', t(BASE, 43000), 'Debug', 'Hook "post-upgrade/smoke-test" started (job api-gateway-smoke-f8k2m)'),
  makeLog('h-047', t(BASE, 44000), 'Info', 'Smoke test: GET /healthz → 200 OK (23ms)'),
  makeLog('h-048', t(BASE, 44500), 'Info', 'Smoke test: GET /api/v1/status → 200 OK (45ms)'),
  makeLog('h-049', t(BASE, 45000), 'Info', 'Smoke test: POST /api/v1/echo → 200 OK (31ms)'),
  makeLog('h-050', t(BASE, 45500), 'Info', 'Post-upgrade hook "smoke-test" completed successfully'),
  makeLog('h-051', t(BASE, 46000), 'Info', 'Release "api-gateway" upgraded to revision 13'),
  makeLog('h-052', t(BASE, 46500), 'Info', 'Helm deploy completed successfully in 46.5s'),
  // system logs interspersed
  makeLog('h-sys-1', t(BASE, 500), 'Trace', 'ECR auth token cached, expires in 11h58m', 'runner', 'system'),
  makeLog('h-sys-2', t(BASE, 8000), 'Debug', 'kube client: PATCH deployments/api-gateway 200 OK (82ms)', 'runner', 'system'),
  makeLog('h-sys-3', t(BASE, 16000), 'Trace', 'kube watcher: deployment api-gateway event MODIFIED', 'runner', 'system'),
  makeLog('h-sys-4', t(BASE, 22000), 'Trace', 'kube watcher: deployment api-gateway event MODIFIED', 'runner', 'system'),
  makeLog('h-sys-5', t(BASE, 28000), 'Trace', 'kube watcher: deployment api-gateway event MODIFIED', 'runner', 'system'),
  makeLog('h-sys-6', t(BASE, 34000), 'Debug', 'kube watcher: pod api-gateway-7b8f4d6c9-q3k9x restart count incremented', 'runner', 'system'),
  makeLog('h-sys-7', t(BASE, 40000), 'Trace', 'kube watcher: deployment api-gateway event MODIFIED', 'runner', 'system'),
].sort((a, b) => a.timestamp.localeCompare(b.timestamp))

const dockerBuildLogs: TOTELLog[] = [
  makeLog('d-001', t(BASE, 0), 'Info', 'Starting Docker build for component "worker"', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-002', t(BASE, 100), 'Info', 'Build context: /workspace/worker (14 files, 2.8 MB)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-003', t(BASE, 300), 'Debug', 'Dockerfile: ./worker/Dockerfile', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-004', t(BASE, 500), 'Debug', 'Build args: GO_VERSION=1.25, ALPINE_VERSION=3.20', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-005', t(BASE, 800), 'Info', '[1/12] FROM docker.io/library/golang:1.25-alpine3.20@sha256:abc123', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-006', t(BASE, 900), 'Debug', '[1/12] CACHED', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-007', t(BASE, 1000), 'Info', '[2/12] RUN apk add --no-cache git ca-certificates tzdata', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-008', t(BASE, 2500), 'Debug', 'fetch https://dl-cdn.alpinelinux.org/alpine/v3.20/main/x86_64/APKINDEX.tar.gz', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-009', t(BASE, 3500), 'Debug', '(1/4) Installing git (2.45.2-r0)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-010', t(BASE, 4200), 'Debug', '(2/4) Installing ca-certificates (20240226-r0)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-011', t(BASE, 4800), 'Debug', '(3/4) Installing tzdata (2024a-r1)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-012', t(BASE, 5200), 'Debug', '(4/4) Installing libcurl (8.8.0-r0)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-013', t(BASE, 5500), 'Info', '[3/12] WORKDIR /app', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-014', t(BASE, 5700), 'Info', '[4/12] COPY go.mod go.sum ./', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-015', t(BASE, 5900), 'Info', '[5/12] RUN go mod download', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-016', t(BASE, 6500), 'Debug', 'go: downloading github.com/labstack/echo/v4 v4.12.0', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-017', t(BASE, 7000), 'Debug', 'go: downloading github.com/jackc/pgx/v5 v5.6.0', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-018', t(BASE, 7500), 'Debug', 'go: downloading go.uber.org/zap v1.27.0', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-019', t(BASE, 8000), 'Debug', 'go: downloading google.golang.org/grpc v1.65.0', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-020', t(BASE, 8500), 'Debug', 'go: downloading github.com/redis/go-redis/v9 v9.5.3', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-021', t(BASE, 9200), 'Debug', 'go: downloading go.opentelemetry.io/otel v1.28.0', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-022', t(BASE, 10000), 'Info', 'go mod download completed (83 modules, 4.2s)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-023', t(BASE, 10200), 'Info', '[6/12] COPY . .', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-024', t(BASE, 10500), 'Info', '[7/12] RUN go build -ldflags="-s -w -X main.version=v1.8.3" -o /bin/worker ./cmd/worker', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-025', t(BASE, 11000), 'Debug', 'compiling package cmd/worker', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-026', t(BASE, 12000), 'Debug', 'compiling package internal/server', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-027', t(BASE, 13000), 'Debug', 'compiling package internal/handlers', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-028', t(BASE, 14000), 'Debug', 'compiling package internal/repository', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-029', t(BASE, 15000), 'Debug', 'compiling package internal/workers', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-030', t(BASE, 16000), 'Warn', 'internal/legacy/compat.go:42:6: exported function LegacyHandler is unused (SA1019)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-031', t(BASE, 16500), 'Warn', 'internal/legacy/compat.go:78:2: deprecated: use internal/handlers.NewRouter instead', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-032', t(BASE, 18000), 'Debug', 'linking /bin/worker', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-033', t(BASE, 19000), 'Info', 'go build completed (8.5s), binary size 18.4 MB', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-034', t(BASE, 19200), 'Info', '[8/12] FROM docker.io/library/alpine:3.20@sha256:def456', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-035', t(BASE, 19400), 'Debug', '[8/12] CACHED', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-036', t(BASE, 19600), 'Info', '[9/12] RUN addgroup -S app && adduser -S app -G app', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-037', t(BASE, 19900), 'Info', '[10/12] COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-038', t(BASE, 20100), 'Info', '[11/12] COPY --from=0 /bin/worker /bin/worker', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-039', t(BASE, 20300), 'Info', '[12/12] USER app', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-040', t(BASE, 20500), 'Info', 'Exporting layers', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-041', t(BASE, 22000), 'Debug', 'Layer sha256:a1b2c3 (4.7 MB) exported', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-042', t(BASE, 23000), 'Debug', 'Layer sha256:d4e5f6 (18.4 MB) exported', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-043', t(BASE, 24000), 'Debug', 'Layer sha256:g7h8i9 (0.3 MB) exported', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-044', t(BASE, 25000), 'Info', 'Image built: sha256:9f8e7d6c5b4a3210fedcba9876543210fedcba9876543210fedcba9876543210', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-045', t(BASE, 25200), 'Info', 'Tagging image as 766121324316.dkr.ecr.us-west-2.amazonaws.com/org123/app456:bld-abc123', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-046', t(BASE, 25500), 'Info', 'Pushing image to ECR', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-047', t(BASE, 26000), 'Debug', 'Pushing layer sha256:a1b2c3 (4.7 MB)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-048', t(BASE, 28000), 'Debug', 'Pushing layer sha256:d4e5f6 (18.4 MB)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-049', t(BASE, 32000), 'Debug', 'Pushing layer sha256:g7h8i9 (0.3 MB)', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-050', t(BASE, 33000), 'Info', 'All layers pushed successfully', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-051', t(BASE, 33500), 'Info', 'Pushing manifest', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-052', t(BASE, 34000), 'Info', 'Image pushed: 766121324316.dkr.ecr.us-west-2.amazonaws.com/org123/app456:bld-abc123', 'runner', 'oteljob', 'docker_build'),
  makeLog('d-053', t(BASE, 34200), 'Info', 'Docker build completed successfully in 34.2s', 'runner', 'oteljob', 'docker_build'),
  // system logs
  makeLog('d-sys-1', t(BASE, 50), 'Trace', 'docker buildx create --use --name nuon-builder', 'runner', 'system', 'docker_build'),
  makeLog('d-sys-2', t(BASE, 150), 'Debug', 'ECR auth: refreshing token for registry 766121324316.dkr.ecr.us-west-2.amazonaws.com', 'runner', 'system', 'docker_build'),
  makeLog('d-sys-3', t(BASE, 250), 'Trace', 'ECR auth token obtained, valid for 12h', 'runner', 'system', 'docker_build'),
  makeLog('d-sys-4', t(BASE, 21000), 'Debug', 'disk usage: build context 2.8 MB, build cache 847 MB, free 14.2 GB', 'runner', 'system', 'docker_build'),
  makeLog('d-sys-5', t(BASE, 27000), 'Trace', 'ECR push: progress 45% (10.4/23.4 MB)', 'runner', 'system', 'docker_build'),
  makeLog('d-sys-6', t(BASE, 30000), 'Trace', 'ECR push: progress 82% (19.2/23.4 MB)', 'runner', 'system', 'docker_build'),
].sort((a, b) => a.timestamp.localeCompare(b.timestamp))

const failedTerraformLogs: TOTELLog[] = [
  makeLog('tf-001', t(BASE, 0), 'Info', 'Starting Terraform plan for component "vpc-network"', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-002', t(BASE, 200), 'Info', 'Terraform v1.9.2 on linux_amd64', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-003', t(BASE, 500), 'Debug', 'Initializing provider plugins...', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-004', t(BASE, 1000), 'Debug', '- Finding hashicorp/aws versions matching "~> 5.60"...', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-005', t(BASE, 2000), 'Debug', '- Installing hashicorp/aws v5.62.0...', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-006', t(BASE, 3500), 'Debug', '- Installed hashicorp/aws v5.62.0 (signed by HashiCorp)', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-007', t(BASE, 4000), 'Info', 'Terraform has been successfully initialized!', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-008', t(BASE, 4500), 'Info', 'Planning with 3 variables set from environment', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-009', t(BASE, 5000), 'Debug', 'aws_vpc.main: Refreshing state... [id=vpc-0a1b2c3d4e5f67890]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-010', t(BASE, 5500), 'Debug', 'aws_subnet.private[0]: Refreshing state... [id=subnet-0f1e2d3c4b5a6f7e8]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-011', t(BASE, 5800), 'Debug', 'aws_subnet.private[1]: Refreshing state... [id=subnet-0a9b8c7d6e5f4a3b2]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-012', t(BASE, 6100), 'Debug', 'aws_subnet.public[0]: Refreshing state... [id=subnet-0c1d2e3f4a5b6c7d8]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-013', t(BASE, 6400), 'Debug', 'aws_subnet.public[1]: Refreshing state... [id=subnet-0e9f8a7b6c5d4e3f2]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-014', t(BASE, 7000), 'Debug', 'aws_nat_gateway.main: Refreshing state... [id=nat-0a1b2c3d4e5f67890]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-015', t(BASE, 7500), 'Debug', 'aws_security_group.eks_cluster: Refreshing state... [id=sg-0f1e2d3c4b5a6f7e8]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-016', t(BASE, 8000), 'Debug', 'aws_security_group.eks_nodes: Refreshing state... [id=sg-0a9b8c7d6e5f4a3b2]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-017', t(BASE, 8500), 'Debug', 'aws_eks_cluster.main: Refreshing state... [id=nuon-prod-us-west-2]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-018', t(BASE, 9000), 'Debug', 'aws_eks_node_group.workers: Refreshing state... [id=nuon-prod-us-west-2:workers]', 'runner', 'oteljob', 'terraform'),
  makeLog('tf-019', t(BASE, 10000), 'Info', 'Terraform plan: 2 to add, 1 to change, 0 to destroy'),
  makeLog('tf-020', t(BASE, 10500), 'Info', '  + aws_subnet.private[2] (new AZ us-west-2c)'),
  makeLog('tf-021', t(BASE, 10800), 'Info', '  + aws_route_table_association.private[2]'),
  makeLog('tf-022', t(BASE, 11100), 'Info', '  ~ aws_eks_node_group.workers (subnet_ids updated)'),
  makeLog('tf-023', t(BASE, 11500), 'Info', 'Applying Terraform plan...'),
  makeLog('tf-024', t(BASE, 12000), 'Info', 'aws_subnet.private[2]: Creating...'),
  makeLog('tf-025', t(BASE, 14000), 'Info', 'aws_subnet.private[2]: Creation complete after 2s [id=subnet-0new123456789abcd]'),
  makeLog('tf-026', t(BASE, 14500), 'Info', 'aws_route_table_association.private[2]: Creating...'),
  makeLog('tf-027', t(BASE, 15500), 'Info', 'aws_route_table_association.private[2]: Creation complete after 1s'),
  makeLog('tf-028', t(BASE, 16000), 'Info', 'aws_eks_node_group.workers: Modifying... [id=nuon-prod-us-west-2:workers]'),
  makeLog('tf-029', t(BASE, 20000), 'Warn', 'aws_eks_node_group.workers: Still modifying... [4s elapsed]'),
  makeLog('tf-030', t(BASE, 30000), 'Warn', 'aws_eks_node_group.workers: Still modifying... [14s elapsed]'),
  makeLog('tf-031', t(BASE, 60000), 'Warn', 'aws_eks_node_group.workers: Still modifying... [44s elapsed]'),
  makeLog('tf-032', t(BASE, 120000), 'Warn', 'aws_eks_node_group.workers: Still modifying... [1m44s elapsed]'),
  makeLog('tf-033', t(BASE, 180000), 'Error', 'aws_eks_node_group.workers: Error updating EKS Node Group (nuon-prod-us-west-2:workers):'),
  makeLog('tf-034', t(BASE, 180100), 'Error', '  InvalidParameterException: Subnet subnet-0new123456789abcd does not have enough available IP addresses'),
  makeLog('tf-035', t(BASE, 180200), 'Error', '  The subnet CIDR block 10.0.8.0/26 provides 59 usable IPs but the node group requires at least 64'),
  makeLog('tf-036', t(BASE, 180500), 'Fatal', 'Terraform apply failed: 1 error occurred'),
  makeLog('tf-037', t(BASE, 180600), 'Fatal', '  * aws_eks_node_group.workers: insufficient IP addresses in subnet-0new123456789abcd'),
  makeLog('tf-038', t(BASE, 181000), 'Error', 'Apply complete! Resources: 2 added, 0 changed, 0 destroyed.'),
  makeLog('tf-039', t(BASE, 181200), 'Error', 'Note: the EKS node group update was rolled back by AWS. The new subnet was created but nodes could not be launched into it.'),
  makeLog('tf-040', t(BASE, 181500), 'Info', 'Terraform state saved to remote backend (S3)'),
  // system logs
  makeLog('tf-sys-1', t(BASE, 100), 'Trace', 'terraform workspace select install-ins123', 'runner', 'system', 'terraform'),
  makeLog('tf-sys-2', t(BASE, 400), 'Debug', 'S3 state lock acquired: terraform-state/vpc-network/ins123.tflock', 'runner', 'system', 'terraform'),
  makeLog('tf-sys-3', t(BASE, 9500), 'Trace', 'AWS API call count: 47 (ec2: 31, eks: 12, iam: 4)', 'runner', 'system', 'terraform'),
  makeLog('tf-sys-4', t(BASE, 90000), 'Debug', 'AWS API throttle: ec2.DescribeSubnets rate limit hit, backing off 2s', 'runner', 'system', 'terraform'),
  makeLog('tf-sys-5', t(BASE, 181800), 'Debug', 'S3 state lock released', 'runner', 'system', 'terraform'),
].sort((a, b) => a.timestamp.localeCompare(b.timestamp))

const StoryWithLogs = ({ logs }: { logs: TOTELLog[] }) => {
  const filters = useLogFilters(logs)
  const { activeLog, handleActiveLog } = useLogPanel(filters.filteredLogs)

  return (
    <LogStreamContext.Provider
      value={{
        logs,
        logStreamId: 'lgs-story',
        isLoading: false,
        error: null,
        connectionState: 'disconnected' as const,
      }}
    >
      <SSELogs
        filteredLogs={filters.filteredLogs ?? []}
        filters={filters}
        activeLog={activeLog}
        handleActiveLog={handleActiveLog}
        isLoading={false}
        isConnected={false}
      />
    </LogStreamContext.Provider>
  )
}

export const HelmDeploy = () => <StoryWithLogs logs={helmDeployLogs} />
export const DockerBuild = () => <StoryWithLogs logs={dockerBuildLogs} />
export const FailedTerraform = () => <StoryWithLogs logs={failedTerraformLogs} />
