import { api } from '@/lib/api'

export type TProcessUptime = {
  process_id: string
  runner_id: string
  type: string
  status: string
  version: string
  started_at: string
  last_heartbeat?: string
  uptime_ms: number
  uptime_str: string
  heartbeats: number
  health_checks: number
  healthy_checks: number
  unhealthy_checks: number
}

export type TJobSummary = {
  total: number
  finished: number
  failed: number
  timed_out: number
  cancelled: number
  other: number
}

export type TUptimeMetrics = {
  effective_window_ms: number
  total_uptime_ms: number
  total_procs: number
  restarts: number
  total_heartbeats: number
  expected_heartbeats: number
  total_health_checks: number
  healthy_checks: number
  unhealthy_checks: number
  expected_health_checks: number
}

export type TInstallUptimeEntry = {
  install_id: string
  install_name: string
  org_id: string
  org_name: string
  runner_created_at?: string
  install_processes: TProcessUptime[]
  mng_processes: TProcessUptime[]
  install_metrics: TUptimeMetrics
  mng_metrics: TUptimeMetrics
  jobs: TJobSummary
}

export const getRunnerUptime = (params?: {
  org_id?: string
  install_name?: string
  label?: string
  window?: string
}) =>
  api<{
    installs: TInstallUptimeEntry[]
    window: string
    since: string
    window_ms: number
    orgs: { id: string; name: string }[]
    label_options: { key: string; values: string[] }[]
  }>({ path: 'runner-uptime', params })
