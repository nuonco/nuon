import { useState, type ReactNode } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { NavLink, Route, Routes } from 'react-router'
import { dispatchRef, getCatalog, getHealth, getRun, getRuns } from './api'
import type { TDrift, TRun } from './types'

const formatDate = (value?: string) =>
  value ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(value)) : '—'

const statusTheme = (status?: string) => {
  const normalized = status?.toLowerCase() ?? ''
  if (['healthy', 'finished', 'succeeded', 'success', 'no drift'].includes(normalized)) return 'success'
  if (['failed', 'error', 'unhealthy', 'drifted'].includes(normalized)) return 'error'
  if (['degraded', 'warning'].includes(normalized)) return 'warn'
  return 'neutral'
}

const Pill = ({ children, theme }: { children: ReactNode; theme?: string }) => (
  <span className={`pill ${theme ?? statusTheme(String(children))}`}>{children}</span>
)

const State = ({ error, loading, empty }: { error?: Error | null; loading?: boolean; empty?: boolean }) => {
  if (loading) return <div className="state">Loading…</div>
  if (error) return <div className="state error-text">{error.message}</div>
  if (empty) return <div className="state">No data found</div>
  return null
}

const Page = ({ title, description, children }: { title: string; description: string; children: ReactNode }) => (
  <main>
    <div className="page-heading">
      <div>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      <span className="polling"><i /> Refreshing every 5 seconds</span>
    </div>
    <section className="card">{children}</section>
  </main>
)

const ComponentsView = () => {
  const health = useQuery({ queryKey: ['health'], queryFn: getHealth })
  const components = health.data?.latest?.components ?? []
  return (
    <Page title="Components" description="Current health reported by the resident runner.">
      <State error={health.error} loading={health.isLoading} empty={!health.isLoading && components.length === 0} />
      {components.length > 0 && (
        <div className="table-wrap"><table><thead><tr><th>Name</th><th>Type</th><th>Health</th></tr></thead><tbody>
          {components.map((component) => <tr key={component.install_component_id ?? component.component_id ?? component.component_name}>
            <td className="strong">{component.component_name ?? component.component_id ?? 'Unnamed component'}</td>
            <td><code>{component.component_type ?? '—'}</code></td>
            <td><Pill>{component.health || 'unknown'}</Pill></td>
          </tr>)}
        </tbody></table></div>
      )}
    </Page>
  )
}

const RefsView = () => {
  const queryClient = useQueryClient()
  const catalog = useQuery({ queryKey: ['catalog'], queryFn: getCatalog })
  const [message, setMessage] = useState('')
  const dispatch = useMutation({
    mutationFn: dispatchRef,
    onSuccess: (data) => {
      setMessage(`Dispatch ${data.dispatch_id} queued`)
      queryClient.invalidateQueries({ queryKey: ['runs'] })
    },
    onError: (error) => setMessage(error.message),
  })
  const refs = catalog.data?.refs ?? []
  return (
    <Page title="Refs" description="Actions, runbooks, and drift checks available in this bundle.">
      {message && <div className={dispatch.isError ? 'notice error-text' : 'notice'}>{message}</div>}
      <State error={catalog.error} loading={catalog.isLoading} empty={!catalog.isLoading && refs.length === 0} />
      {refs.length > 0 && <div className="table-wrap"><table><thead><tr><th>Name</th><th>Kind</th><th>Component</th><th>Cron</th><th /></tr></thead><tbody>
        {refs.map((ref) => <tr key={ref.id}><td><div className="strong">{ref.name}</div><code className="subtle">{ref.id}</code></td><td><Pill theme="brand">{ref.kind}</Pill></td><td>{ref.component ?? '—'}</td><td><code>{ref.cron_schedule ?? '—'}</code></td><td className="actions"><button disabled={dispatch.isPending} onClick={() => dispatch.mutate(ref.id)}>{dispatch.isPending && dispatch.variables === ref.id ? 'Queuing…' : 'Run'}</button></td></tr>)}
      </tbody></table></div>}
    </Page>
  )
}

const Drift = ({ drift }: { drift: TDrift }) => (
  <div className="drift">
    <div><Pill theme={drift.drifted ? 'error' : 'success'}>{drift.drifted ? 'DRIFTED' : 'no drift'}</Pill>{drift.summary && <span>{drift.summary}</span>}</div>
    <dl><div><dt>Resource changes</dt><dd>{drift.resource_changes}</dd></div><div><dt>Output changes</dt><dd>{drift.output_changes}</dd></div><div><dt>Resource drift</dt><dd>{drift.resource_drift}</dd></div></dl>
  </div>
)

const RunDetail = ({ runId, initial }: { runId: string; initial: TRun }) => {
  const detail = useQuery({ queryKey: ['run', runId], queryFn: () => getRun(runId), initialData: initial })
  const run = detail.data
  return <div className="run-detail">
    {run.error && <div className="error-box"><b>Run error</b><pre>{run.error}</pre></div>}
    <h3>Steps</h3>
    {(run.steps ?? []).length === 0 ? <div className="state">No steps recorded</div> : <div className="steps">
      {(run.steps ?? []).map((step) => <article className="step" key={step.id}>
        <div className="step-heading"><div><span className="strong">{step.name}</span><code>{step.kind}</code></div><Pill>{step.status}</Pill></div>
        {step.job_id && <div className="job"><span>Job ID</span><code>{step.job_id}</code></div>}
        {step.error && <pre className="error-box">{step.error}</pre>}
        {step.drift && <Drift drift={step.drift} />}
      </article>)}
    </div>}
  </div>
}

const RunsView = () => {
  const runs = useQuery({ queryKey: ['runs'], queryFn: getRuns })
  const [expanded, setExpanded] = useState<string | null>(null)
  const sorted = [...(runs.data ?? [])].sort((a, b) => Date.parse(b.started_at) - Date.parse(a.started_at))
  return (
    <Page title="Runs" description="Execution history, newest first.">
      <State error={runs.error} loading={runs.isLoading} empty={!runs.isLoading && sorted.length === 0} />
      {sorted.length > 0 && <div className="runs">{sorted.map((run) => {
        const open = expanded === run.run_id
        return <article className="run" key={run.run_id}>
          <button className="run-summary" aria-expanded={open} onClick={() => setExpanded(open ? null : run.run_id)}>
            <span className="chevron">›</span><span><b>{run.ref_name || run.ref_id}</b><code>{run.run_id}</code></span><span><small>Kind</small>{run.ref_kind}</span><span><small>Source</small>{run.source}</span><span><small>Started</small>{formatDate(run.started_at)}</span><Pill>{run.status}</Pill>
          </button>
          {open && <RunDetail runId={run.run_id} initial={run} />}
        </article>
      })}</div>}
    </Page>
  )
}

const Layout = () => {
  const catalog = useQuery({ queryKey: ['catalog'], queryFn: getCatalog })
  return <div className="shell">
    <header><div className="brand"><span className="mark">N</span><span>Bundle portal</span></div><nav><NavLink end to="/">Components</NavLink><NavLink to="/refs">Refs</NavLink><NavLink to="/runs">Runs</NavLink></nav><div className="identity"><span><small>Deployment / install</small><code>{catalog.data?.deployment_id ?? 'Loading…'}</code></span><span><small>Bundle digest</small><code title={catalog.data?.bundle_digest}>{catalog.data?.bundle_digest ?? 'Loading…'}</code></span></div></header>
    <Routes><Route path="/" element={<ComponentsView />} /><Route path="/refs" element={<RefsView />} /><Route path="/runs" element={<RunsView />} /></Routes>
  </div>
}

export const App = () => <Layout />
