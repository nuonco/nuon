import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import {
  getRunnerUptime,
  type TInstallUptimeEntry,
  type TProcessUptime,
  type TJobSummary,
  type TUptimeMetrics,
} from '@/lib/admin-api'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { Badge } from '@/components/common/Badge'

const WINDOWS = [
  { value: 'today', label: 'Today' },
  { value: 'week', label: 'Past week' },
  { value: 'month', label: 'Past month' },
  { value: 'quarter', label: 'Past quarter' },
] as const

// --- Pie chart (SVG donut) ---
const PieChart = ({ value, max, label, color }: { value: number; max: number; label: string; color: string }) => {
  const pct = max > 0 ? Math.min(value / max, 1) : 0
  const r = 16
  const circ = 2 * Math.PI * r
  const filled = circ * pct
  const gap = circ - filled
  return (
    <div className="flex flex-col items-center gap-0.5">
      <svg width="44" height="44" viewBox="0 0 44 44">
        <circle cx="22" cy="22" r={r} fill="none" className="stroke-gray-200 dark:stroke-gray-700" strokeWidth="5" />
        <circle
          cx="22" cy="22" r={r} fill="none"
          stroke={color} strokeWidth="5"
          strokeDasharray={`${filled} ${gap}`}
          strokeDashoffset={circ / 4}
          strokeLinecap="round"
        />
        <text x="22" y="23" textAnchor="middle" dominantBaseline="middle" className="text-[9px] font-semibold fill-gray-700 dark:fill-gray-200">
          {(pct * 100).toFixed(0)}%
        </text>
      </svg>
      <span className="text-[9px] text-gray-500 dark:text-gray-400 text-center leading-tight">{label}</span>
      <span className="text-[9px] text-gray-400 dark:text-gray-500">{value}/{max}</span>
    </div>
  )
}

const statusColor = (s: string) => {
  switch (s) {
    case 'active': return 'bg-green-100 text-green-700'
    case 'inactive': case 'offline': return 'bg-gray-100 text-gray-600'
    case 'error': return 'bg-red-100 text-red-700'
    default: return 'bg-yellow-100 text-yellow-700'
  }
}

const processTypeLabel = (t: string) => {
  switch (t) {
    case 'build': return 'mng'
    case 'install': return 'install'
    default: return t
  }
}

const JobBar = ({ jobs }: { jobs: TJobSummary }) => {
  if (jobs.total === 0) return <span className="text-xs text-gray-400">No jobs</span>
  const segments = [
    { count: jobs.finished, color: 'bg-green-500', label: 'finished' },
    { count: jobs.failed, color: 'bg-red-500', label: 'failed' },
    { count: jobs.timed_out, color: 'bg-yellow-500', label: 'timed out' },
    { count: jobs.cancelled, color: 'bg-gray-400', label: 'cancelled' },
    { count: jobs.other, color: 'bg-blue-300', label: 'other' },
  ]
  return (
    <div>
      <div className="flex h-2 w-full rounded-full overflow-hidden bg-gray-100">
        {segments.map((seg) =>
          seg.count > 0 ? (
            <div
              key={seg.label}
              className={`${seg.color} h-full`}
              style={{ width: `${(seg.count / jobs.total) * 100}%` }}
              title={`${seg.label}: ${seg.count}`}
            />
          ) : null
        )}
      </div>
      <div className="mt-1 flex gap-3 text-[10px] text-gray-500">
        <span>{jobs.total} total</span>
        <span className="text-green-600">{jobs.finished} ok</span>
        {jobs.failed > 0 && <span className="text-red-600">{jobs.failed} fail</span>}
        {jobs.timed_out > 0 && <span className="text-yellow-600">{jobs.timed_out} timeout</span>}
      </div>
    </div>
  )
}

const MetricsPies = ({ m }: { m: TUptimeMetrics }) => {
  const effectiveWindowS = Math.round(m.effective_window_ms / 1000)
  const uptimeS = Math.round(m.total_uptime_ms / 1000)

  return (
    <div className="flex gap-4 flex-wrap">
      <PieChart value={uptimeS} max={effectiveWindowS} label="Uptime" color="#22c55e" />
      <div className="flex flex-col items-center gap-0.5 justify-center">
        <span className="text-lg font-semibold text-gray-700 dark:text-gray-200">{m.total_procs}</span>
        <span className="text-[9px] text-gray-500 dark:text-gray-400">processes</span>
        {m.restarts > 0 && <span className="text-[9px] text-orange-500 dark:text-orange-400">{m.restarts} restart{m.restarts > 1 ? 's' : ''}</span>}
      </div>
      {m.expected_heartbeats > 0 && (
        <PieChart value={Number(m.total_heartbeats)} max={Number(m.expected_heartbeats)} label="Heartbeats" color="#a855f7" />
      )}
      {m.expected_health_checks > 0 && (
        <div className="flex flex-col items-center gap-0.5">
          <PieChart
            value={Number(m.total_health_checks)}
            max={Number(m.expected_health_checks)}
            label="HC reported"
            color="#f59e0b"
          />
          {m.total_health_checks > 0 && (
            <div className="flex gap-1.5 text-[9px]">
              <span className="text-green-600">{m.healthy_checks} ok</span>
              {m.unhealthy_checks > 0 && <span className="text-red-600">{m.unhealthy_checks} bad</span>}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

const ProcessBlock = ({ procs, metrics, label, color }: {
  procs: TProcessUptime[]
  metrics: TUptimeMetrics
  label: string
  color: string
}) => {
  return (
    <div className={`rounded-md border p-3 ${color}`}>
      <h4 className="text-xs font-semibold text-gray-700 mb-2">{label}</h4>
      <MetricsPies m={metrics} />
      <div className="mt-2">
        <h5 className="text-[10px] font-semibold text-gray-500 mb-1">Processes</h5>
        {procs.length > 0 ? (
          <div className="divide-y divide-gray-50">
            {procs.map((p) => (
              <div key={p.process_id} className="flex items-center gap-3 py-1.5 text-xs">
                <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium ${statusColor(p.status)}`}>
                  {p.status}
                </span>
                <Badge>{processTypeLabel(p.type)}</Badge>
                <span className="font-mono text-gray-500 w-20" title="Process uptime">{p.uptime_str || '—'}</span>
                {p.heartbeats > 0 && <span className="text-gray-400" title="Heartbeats">{p.heartbeats} hb</span>}
                {p.health_checks > 0 && (
                  <span className="text-gray-400" title={`${p.healthy_checks} ok / ${p.unhealthy_checks} bad`}>
                    {p.health_checks} hc
                    {p.unhealthy_checks > 0 && <span className="text-red-500 ml-0.5">({p.unhealthy_checks} bad)</span>}
                  </span>
                )}
                {p.last_heartbeat && (
                  <span className="text-[10px] text-gray-400" title="Last heartbeat">
                    last hb {new Date(p.last_heartbeat).toLocaleTimeString()}
                  </span>
                )}
                <span className="text-gray-400 font-mono text-[10px] ml-auto">{p.version}</span>
              </div>
            ))}
          </div>
        ) : (
          <span className="text-xs text-gray-400">No processes</span>
        )}
      </div>
    </div>
  )
}

export const RunnerUptime = () => {
  const [orgId, setOrgId] = useState('')
  const [installName, setInstallName] = useState('')
  const [labelKey, setLabelKey] = useState('')
  const [labelValue, setLabelValue] = useState('')
  const [window, setWindow] = useState('today')
  const [expanded, setExpanded] = useState<string | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['runner-uptime', orgId, installName, labelKey, labelValue, window],
    queryFn: () => {
      let label: string | undefined
      if (labelKey && labelValue) {
        label = `${labelKey}:${labelValue}`
      } else if (labelKey) {
        label = labelKey
      }
      return getRunnerUptime({
        org_id: orgId || undefined,
        install_name: installName || undefined,
        label,
        window,
      })
    },
    refetchInterval: 30000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load'} />

  const installs = data?.installs || []
  const orgs = data?.orgs || []
  const labelOptions = data?.label_options || []
  const selectedLabelOption = labelOptions.find((l) => l.key === labelKey)
  const since = data?.since || ''

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="page-heading">Runner uptime</h1>
        {since && (
          <span className="text-xs text-gray-400">Since {new Date(since).toLocaleString()}</span>
        )}
      </div>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:flex-wrap">
        <select
          value={orgId}
          onChange={(e) => setOrgId(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
        >
          <option value="">All orgs</option>
          {orgs.map((o) => (
            <option key={o.id} value={o.id}>{o.name}</option>
          ))}
        </select>
        <div className="w-full sm:w-48">
          <SearchInput value={installName} onChange={setInstallName} placeholder="Install name..." />
        </div>
        <select
          value={labelKey}
          onChange={(e) => { setLabelKey(e.target.value); setLabelValue('') }}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
        >
          <option value="">All labels</option>
          {labelOptions.map((l) => (
            <option key={l.key} value={l.key}>{l.key}</option>
          ))}
        </select>
        {labelKey && selectedLabelOption && selectedLabelOption.values.length > 0 && (
          <select
            value={labelValue}
            onChange={(e) => setLabelValue(e.target.value)}
            className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
          >
            <option value="">Any value</option>
            {selectedLabelOption.values.map((v) => (
              <option key={v} value={v}>{v}</option>
            ))}
          </select>
        )}
        <select
          value={window}
          onChange={(e) => setWindow(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300"
        >
          {WINDOWS.map((w) => (
            <option key={w.value} value={w.value}>{w.label}</option>
          ))}
        </select>
      </div>

      <div className="mt-4 space-y-2">
        {installs.map((entry: TInstallUptimeEntry) => {
          const isExpanded = expanded === entry.install_id
          const activeInstall = entry.install_metrics.active_procs
          const activeMng = entry.mng_metrics.active_procs
          const totalJobs = entry.jobs.total
          const finishedJobs = entry.jobs.finished
          const successRate = totalJobs > 0 ? ((finishedJobs / totalJobs) * 100).toFixed(1) : null

          return (
            <div key={entry.install_id} className="rounded-md border border-gray-200 bg-white">
              <button
                onClick={() => setExpanded(isExpanded ? null : entry.install_id)}
                className="flex w-full items-center gap-3 px-4 py-3 text-left text-sm hover:bg-gray-50"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-gray-900">{entry.install_name || '(unnamed)'}</span>
                  </div>
                  <div className="text-xs text-gray-400 mt-0.5">{entry.org_name}</div>
                </div>
                <div className="flex items-center gap-5 flex-shrink-0 text-xs">
                  <div className="text-center">
                    <div className="font-semibold text-blue-600">{activeInstall}</div>
                    <div className="text-[10px] text-gray-400">install</div>
                  </div>
                  <div className="text-center">
                    <div className="font-semibold text-purple-600">{activeMng}</div>
                    <div className="text-[10px] text-gray-400">mng</div>
                  </div>
                  <div className="text-center">
                    <div className="font-semibold text-gray-700">{totalJobs}</div>
                    <div className="text-[10px] text-gray-400">jobs</div>
                  </div>
                  {successRate && (
                    <div className="text-center">
                      <div className={`font-semibold ${Number(successRate) >= 95 ? 'text-green-600' : Number(successRate) >= 80 ? 'text-yellow-600' : 'text-red-600'}`}>
                        {successRate}%
                      </div>
                      <div className="text-[10px] text-gray-400">success</div>
                    </div>
                  )}
                </div>
              </button>
              {isExpanded && (
                <div className="border-t border-gray-100 px-4 py-3 space-y-4">
                  {entry.runner_created_at && since && new Date(entry.runner_created_at).getTime() > new Date(since).getTime() && (
                    <div className="rounded bg-yellow-50 border border-yellow-200 px-3 py-1.5 text-xs text-yellow-700">
                      Runner created after window began ({new Date(entry.runner_created_at).toLocaleString()}) — stats reflect partial coverage
                    </div>
                  )}
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                    <ProcessBlock
                      procs={entry.install_processes}
                      metrics={entry.install_metrics}
                      label="Install process"
                      color="border-blue-100 bg-blue-50/30"
                    />
                    <ProcessBlock
                      procs={entry.mng_processes}
                      metrics={entry.mng_metrics}
                      label="Management process (mng)"
                      color="border-purple-100 bg-purple-50/30"
                    />
                  </div>
                  <div>
                    <h4 className="text-xs font-semibold text-gray-700 mb-1">Runner jobs (combined)</h4>
                    <JobBar jobs={entry.jobs} />
                  </div>
                  <div className="text-[10px] text-gray-400 font-mono">
                    install: {entry.install_id} | org: {entry.org_id}
                  </div>
                </div>
              )}
            </div>
          )
        })}
        {installs.length === 0 && (
          <div className="text-center text-gray-500 py-8 text-sm">No installs found</div>
        )}
      </div>
    </div>
  )
}
