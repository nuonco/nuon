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
import { useLogFilters } from '@/hooks/use-log-filters'
import type { TOTELLog, TLogStream } from '@/types'

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
      'runner_job.id': 'job3m9xkr7fvp2qtw4ycz1nbah',
      'runner_job.type': 'helm-deploy',
      'runner_job_execution.id': 'runhx5kcw8fvp3qtm2ycz9nbaj',
      'runner_job_execution_step.name': 'helm-upgrade',
      'nuon.tool': 'helm',
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
      'log_stream.id': 'lgsf8k2m4npq1rvx3wtz6yba7',
      'runner_job.id': 'job3m9xkr7fvp2qtw4ycz1nbah',
      'runner_job.type': 'helm-deploy',
      'runner_job_execution.id': 'runhx5kcw8fvp3qtm2ycz9nbaj',
      'runner_job_execution_step.name': 'health-check',
      'nuon.tool': 'helm',
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
    service_name: 'runner.actions.idle',
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
      'nuon.tool': 'terraform',
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
    service_name: 'runner.actions.ddl',
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
      'nuon.tool': 'terraform',
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
    service_name: 'runner.actions.command',
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
      'nuon.tool': 'terraform',
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
    service_name: 'runner.actions.dimngest-operator',
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
    service_name: 'runner.actions.kind',
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
      'nuon.tool': 'kustomize',
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
  {
    id: 'log-11',
    timestamp: '2024-01-15T10:30:15.111Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:15.111',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Authenticating to ECR registry 766121324316.dkr.ecr.us-west-2.amazonaws.com',
    service_name: 'runner',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6',
    span_id: 'a1b2c3d4e5f6a7b8',
    log_attributes: { 'nuon.tool': 'helm' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-12',
    timestamp: '2024-01-15T10:30:16.222Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:16.222',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Pulling chart my-app v2.4.1 from oci://766121324316.dkr.ecr.us-west-2.amazonaws.com/charts',
    service_name: 'runner.actions.helm-pull',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6',
    span_id: 'b2c3d4e5f6a7b8c9',
    log_attributes: { 'nuon.tool': 'helm' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-13',
    timestamp: '2024-01-15T10:30:17.333Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:17.333',
    severity_number: 13,
    severity_text: 'Warn',
    body: 'Helm values override: key "replicaCount" changed from 2 to 3 via install config',
    service_name: 'runner.actions.helm-values',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6',
    span_id: 'c3d4e5f6a7b8c9d0',
    log_attributes: { 'nuon.tool': 'helm' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-14',
    timestamp: '2024-01-15T10:30:18.444Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:18.444',
    severity_number: 9,
    severity_text: 'Info',
    body: 'terraform init -backend-config="bucket=nuon-tf-state-us-west-2" -backend-config="key=install/abcdef/stack.tfstate"',
    service_name: 'runner.actions.terraform-init',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
    span_id: 'd4e5f6a7b8c9d0e1',
    log_attributes: { 'nuon.tool': 'terraform' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-15',
    timestamp: '2024-01-15T10:30:19.555Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:19.555',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Initializing provider plugins... Finding hashicorp/aws versions matching "~> 5.0"...',
    service_name: 'runner.actions.terraform-init',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
    span_id: 'e5f6a7b8c9d0e1f2',
    log_attributes: { 'nuon.tool': 'terraform' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-16',
    timestamp: '2024-01-15T10:30:20.666Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:20.666',
    severity_number: 17,
    severity_text: 'Error',
    body: 'Error: creating EKS Node Group (install-us-east-1:workers): operation error EKS: CreateNodegroup, exceeded max retries',
    service_name: 'runner.actions.terraform-apply',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
    span_id: 'f6a7b8c9d0e1f2a3',
    log_attributes: { 'nuon.tool': 'terraform' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-17',
    timestamp: '2024-01-15T10:30:21.777Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:21.777',
    severity_number: 9,
    severity_text: 'Info',
    body: 'kubectl apply -f manifest.yaml --namespace=production --server-side',
    service_name: 'runner.actions.kubectl-apply',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'd3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8',
    span_id: 'a7b8c9d0e1f2a3b4',
    log_attributes: { 'nuon.tool': 'kustomize' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-18',
    timestamp: '2024-01-15T10:30:22.888Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:22.888',
    severity_number: 9,
    severity_text: 'Info',
    body: 'deployment.apps/web-frontend configured (server-side apply)',
    service_name: 'runner.actions.kubectl-apply',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'd3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8',
    span_id: 'b8c9d0e1f2a3b4c5',
    log_attributes: { 'nuon.tool': 'kustomize' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-19',
    timestamp: '2024-01-15T10:30:23.999Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:23.999',
    severity_number: 5,
    severity_text: 'Debug',
    body: 'Heartbeat sent: runner=rnr4x8gkm2vp7qtw1ycz5nbjh uptime=4h32m cpu=12% mem=341Mi',
    service_name: 'runner',
    scope_name: 'system',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9',
    span_id: 'c9d0e1f2a3b4c5d6',
    log_attributes: {},
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'system' },
  },
  {
    id: 'log-20',
    timestamp: '2024-01-15T10:30:25.100Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:25.100',
    severity_number: 13,
    severity_text: 'Warn',
    body: 'Secret sync: 2 of 5 secrets not found in AWS Secrets Manager — using fallback values',
    service_name: 'runner.actions.secret-sync',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9',
    span_id: 'd0e1f2a3b4c5d6e7',
    log_attributes: {},
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-21',
    timestamp: '2024-01-15T10:30:26.200Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:26.200',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Job completed successfully: type=helm-deploy duration=47s release=my-app namespace=default',
    service_name: 'runner.actions.helm-deploy',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6',
    span_id: 'e1f2a3b4c5d6e7f8',
    log_attributes: { 'nuon.tool': 'helm' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-22',
    timestamp: '2024-01-15T10:30:27.300Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:27.300',
    severity_number: 17,
    severity_text: 'Error',
    body: 'Readiness probe failed: HTTP GET http://10.0.3.47:8080/healthz returned 503 — container web-frontend not ready after 120s',
    service_name: 'runner.actions.health-check',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'd3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8',
    span_id: 'f2a3b4c5d6e7f8a9',
    log_attributes: { 'nuon.tool': 'kustomize' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-23',
    timestamp: '2024-01-15T10:30:28.400Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:28.400',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Workspace lock acquired: workspace=install-us-east-1-stack lock_id=lk_8f3m2n owner=runner-7b8f4d6c9-xk2m4',
    service_name: 'runner.actions.terraform-lock',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
    span_id: 'a3b4c5d6e7f8a9b0',
    log_attributes: { 'nuon.tool': 'terraform' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-24',
    timestamp: '2024-01-15T10:30:29.500Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:29.500',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Apply complete! Resources: 3 added, 1 changed, 0 destroyed.',
    service_name: 'runner.actions.terraform-apply',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
    span_id: 'b4c5d6e7f8a9b0c1',
    log_attributes: { 'nuon.tool': 'terraform' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-25',
    timestamp: '2024-01-15T10:30:30.600Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:30.600',
    severity_number: 1,
    severity_text: 'Trace',
    body: 'gRPC ctl-api.RunnerService/Heartbeat 200 2.1ms runner_id=rnr4x8gkm2vp7qtw1ycz5nbjh',
    service_name: 'runner',
    scope_name: 'system',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0',
    span_id: 'c5d6e7f8a9b0c1d2',
    log_attributes: {},
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'system' },
  },
  {
    id: 'log-26',
    timestamp: '2024-01-15T10:30:31.700Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:31.700',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Workspace lock released: workspace=install-us-east-1-stack lock_id=lk_8f3m2n held_for=63s',
    service_name: 'runner.actions.terraform-lock',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
    span_id: 'd6e7f8a9b0c1d2e3',
    log_attributes: { 'nuon.tool': 'terraform' },
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-27',
    timestamp: '2024-01-15T10:30:32.800Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:32.800',
    severity_number: 13,
    severity_text: 'Warn',
    body: 'Image pull backoff: image=766121324316.dkr.ecr.us-west-2.amazonaws.com/app:v1.2.3 reason=ErrImagePull retries=3',
    service_name: 'runner.actions.image-pull',
    scope_name: 'oteljob',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1',
    span_id: 'e7f8a9b0c1d2e3f4',
    log_attributes: {},
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'oteljob' },
  },
  {
    id: 'log-28',
    timestamp: '2024-01-15T10:30:33.900Z',
    timestamp_date: '2024-01-15',
    timestamp_time: '10:30:33.900',
    severity_number: 9,
    severity_text: 'Info',
    body: 'Runner process started: pid=1 version=v0.14.82 go=1.25.0 arch=amd64',
    service_name: 'runner',
    scope_name: 'system',
    log_stream_id: 'lgsf8k2m4npq1rvx3wtz6yba7',
    org_id: 'orgrok933tcyzji01s7us3aeo3',
    runner_id: 'rnr4x8gkm2vp7qtw1ycz5nbjh',
    runner_group_id: 'rgrpnk3m7fvx9qtz2wyc4abj6',
    trace_id: 'b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2',
    span_id: 'f8a9b0c1d2e3f4a5',
    log_attributes: {},
    resource_attributes: { 'service.name': 'runner' },
    scope_attributes: { 'otel.scope.name': 'system' },
  },
] as TOTELLog[]

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
        loadMore={noop}
        hasMore={false}
        isLoading={false}
        isStreamOpen={false}
      />
    </Providers>
  )
}

export const WithLoadMore = () => {
  const filters = useLogFilters(mockLogs)
  const { activeLog, handleActiveLog } = useLogPanel(filters.filteredLogs)

  return (
    <Providers>
      <SSELogs
        filteredLogs={filters.filteredLogs ?? []}
        filters={filters}
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

export const Loading = () => {
  const filters = useLogFilters([])

  return (
    <Providers>
      <SSELogs
        filteredLogs={[]}
        filters={filters}
        activeLog={undefined}
        handleActiveLog={noop}
        loadMore={noop}
        hasMore={false}
        isLoading={true}
        isStreamOpen={false}
      />
    </Providers>
  )
}

export const Skeleton = () => <LogsSkeleton />
