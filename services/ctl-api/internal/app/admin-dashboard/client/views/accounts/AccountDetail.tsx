import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getAccountDetail, getAccountInstalls, getAccountAuditLogs } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return String(s.status)
  return String(s)
}

export const AccountDetail = () => {
  const { id } = useParams<{ id: string }>()
  const [installsPage, setInstallsPage] = useState(1)
  const [auditPage, setAuditPage] = useState(1)
  const [auditEntityType, setAuditEntityType] = useState('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['account', id],
    queryFn: () => getAccountDetail(id!),
    enabled: !!id,
  })

  const { data: installsData } = useQuery({
    queryKey: ['account-installs', id, installsPage],
    queryFn: () => getAccountInstalls(id!, { page: installsPage }),
    enabled: !!id,
  })

  const { data: auditData } = useQuery({
    queryKey: ['account-audit', id, auditPage, auditEntityType, startDate, endDate],
    queryFn: () =>
      getAccountAuditLogs(id!, {
        page: auditPage,
        entity_type: auditEntityType || undefined,
        start_date: startDate || undefined,
        end_date: endDate || undefined,
      }),
    enabled: !!id,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load account'} />
  if (!data) return null

  const { account, apps = [], installs: accountInstalls = [] } = data

  // Derive unique orgs from roles
  const orgsMap = new Map<string, { org_id: string; org_name: string; role_type: string; created_at: string }>()
  for (const role of account.roles || []) {
    if (role.org_id && !orgsMap.has(role.org_id)) {
      orgsMap.set(role.org_id, {
        org_id: role.org_id,
        org_name: role.org?.name || truncateId(role.org_id),
        role_type: role.role_type,
        created_at: role.created_at,
      })
    }
  }
  const orgsList = Array.from(orgsMap.values())

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <Link to="/accounts" className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">Accounts</Link>
        <span>/</span>
        <span className="text-gray-900 dark:text-gray-100">{account.email}</span>
      </nav>

      {/* Page Heading */}
      <div className="page-heading">
        <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">{account.email}</h1>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400 font-mono">{account.id}</p>
        {account.subject && (
          <p className="mt-0.5 text-xs text-gray-400 dark:text-gray-500 font-mono">Subject: {account.subject}</p>
        )}
        <div className="mt-2 flex items-center gap-2">
          <Badge>{account.account_type}</Badge>
        </div>
        <div className="mt-1 flex items-center gap-4 text-xs text-gray-400 dark:text-gray-500">
          <span>Created {formatDate(account.created_at)}</span>
          <span>Updated {formatDate(account.updated_at)}</span>
        </div>
      </div>

      {/* User Journey */}
      {Array.isArray(account.user_journeys) && account.user_journeys.length > 0 && (
        <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-4">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">User Journey</h2>
          <p className="mt-0.5 text-xs text-gray-500 dark:text-gray-400">Onboarding progress and completion tracking</p>
          <div className="mt-3 space-y-4">
            {account.user_journeys.map((journey: any, ji: number) => {
              const steps: any[] = journey.steps || []
              const completed = steps.filter((s) => s.complete).length
              const total = steps.length
              const pct = total > 0 ? Math.round((completed / total) * 100) : 0
              return (
                <div key={ji} className={ji > 0 ? 'border-t border-gray-100 dark:border-gray-800 pt-3' : ''}>
                  <div className="flex items-center justify-between gap-3">
                    <h3 className="text-xs font-semibold text-gray-700 dark:text-gray-300">{journey.title || journey.name}</h3>
                    <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">{completed} / {total}</span>
                  </div>
                  <div className="mt-1.5 h-1.5 w-full rounded bg-gray-100 dark:bg-gray-800">
                    <div
                      className="h-1.5 rounded bg-primary-500"
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <ul className="mt-3 space-y-1.5">
                    {steps.map((step: any) => (
                      <li key={step.name} className="flex items-start gap-2 text-xs">
                        <span
                          className={`mt-0.5 inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-full text-[10px] font-semibold ${
                            step.complete
                              ? 'bg-primary-500 text-white'
                              : 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500'
                          }`}
                        >
                          {step.complete ? '✓' : '○'}
                        </span>
                        <div className="flex-1">
                          <div className={step.complete ? 'text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400'}>
                            {step.title || step.name}
                          </div>
                          {step.complete && step.completed_at && (
                            <div className="text-[11px] text-gray-400 dark:text-gray-500">
                              Completed {formatDate(step.completed_at)}
                              {step.completion_method && <> · via {step.completion_method}</>}
                            </div>
                          )}
                          {step.metadata && Object.keys(step.metadata).length > 0 && (
                            <div className="mt-0.5 flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-gray-500 dark:text-gray-400">
                              {Object.entries(step.metadata).map(([k, v]) => (
                                <span key={k} className="font-mono">
                                  <span className="text-gray-400 dark:text-gray-500">{k}:</span> {String(v)}
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Organizations */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Organizations</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Org Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Role Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {orgsList.map((o) => (
                <tr key={o.org_id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/orgs/${o.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-medium">
                      {o.org_name}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge>{o.role_type}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(o.created_at)}</td>
                </tr>
              ))}
              {orgsList.length === 0 && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No organizations</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Roles */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Roles</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Role Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Org</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(account.roles || []).map((role: any) => (
                <tr key={role.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge>{role.role_type}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/orgs/${role.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                      {role.org?.name || truncateId(role.org_id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(role.created_at)}</td>
                </tr>
              ))}
              {(!account.roles || account.roles.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No roles</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Apps */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Apps</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Organization</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Configs</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {apps.map((app: any) => (
                <tr key={app.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-medium text-gray-900 dark:text-gray-100">{app.name}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 font-mono">{truncateId(app.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {app.org_id ? (
                      <Link to={`/orgs/${app.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                        {app.org?.name || truncateId(app.org_id)}
                      </Link>
                    ) : '-'}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    {app.status ? (
                      <Badge variant="status" status={getStatus(app.status)}>{getStatus(app.status)}</Badge>
                    ) : '-'}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{app.config_count ?? '-'}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(app.created_at)}</td>
                </tr>
              ))}
              {apps.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No apps</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Installs */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Installs</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Organization</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">App</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Runner</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Sandbox</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Component</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(installsData?.installs || accountInstalls).map((install: any) => {
                const isDeleted = install.deleted_at && install.deleted_at > 0
                return (
                  <tr key={install.id} className={isDeleted ? 'opacity-50' : ''}>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <div className="flex items-center gap-1">
                        <Link to={`/installs/${install.id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200 font-medium">
                          {install.name || truncateId(install.id)}
                        </Link>
                        {isDeleted && <Badge variant="status" status="error">Deleted</Badge>}
                      </div>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      {install.org_id ? (
                        <Link to={`/orgs/${install.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                          {install.org?.name || truncateId(install.org_id)}
                        </Link>
                      ) : '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                      {install.app?.name || truncateId(install.app_id)}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Badge variant="status" status={getStatus(install.status)}>{getStatus(install.status)}</Badge>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      {install.runner_status ? (
                        <Badge variant="status" status={getStatus(install.runner_status)}>{getStatus(install.runner_status)}</Badge>
                      ) : '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      {install.sandbox_status ? (
                        <Badge variant="status" status={getStatus(install.sandbox_status)}>{getStatus(install.sandbox_status)}</Badge>
                      ) : '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      {install.composite_component_status ? (
                        <Badge variant="status" status={getStatus(install.composite_component_status)}>{getStatus(install.composite_component_status)}</Badge>
                      ) : '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(install.created_at)}</td>
                  </tr>
                )
              })}
              {(installsData?.installs || accountInstalls).length === 0 && (
                <tr>
                  <td colSpan={8} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No installs</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {installsData && (
          <Pagination page={installsPage} totalPages={installsData.total_pages} onPageChange={setInstallsPage} />
        )}
      </div>

      {/* Audit Logs */}
      <div className="table-card p-4">
        <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100">Audit Logs</h2>
        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
          <select
            value={auditEntityType}
            onChange={(e) => { setAuditEntityType(e.target.value); setAuditPage(1) }}
            className="block w-48 rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          >
            <option value="">All types</option>
            <option value="app">App</option>
            <option value="workflow">Workflow</option>
            <option value="runner_job">Runner Job</option>
            <option value="org">Org</option>
            <option value="app_sync">App Sync</option>
          </select>
          <input
            type="date"
            value={startDate}
            onChange={(e) => { setStartDate(e.target.value); setAuditPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          />
          <input
            type="date"
            value={endDate}
            onChange={(e) => { setEndDate(e.target.value); setAuditPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 dark:text-gray-100 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700 focus:ring-2 focus:ring-primary-600 dark:focus:ring-primary-500"
          />
        </div>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
            <thead className="">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Entity Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Entity ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Org</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Description</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
              {(auditData?.audit_logs || []).map((entry: any, idx: number) => {
                const entityLink = entry.entity_type === 'workflow'
                  ? `/workflows/${entry.entity_id}`
                  : entry.entity_type === 'org'
                    ? `/orgs/${entry.entity_id}`
                    : null
                return (
                  <tr key={entry.entity_id + '-' + idx}>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      <Badge>{entry.entity_type}</Badge>
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900 dark:text-gray-100">{entry.entity_name || '-'}</td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm font-mono">
                      {entityLink ? (
                        <Link to={entityLink} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                          {truncateId(entry.entity_id)}
                        </Link>
                      ) : (
                        <span className="text-gray-500 dark:text-gray-400">{truncateId(entry.entity_id)}</span>
                      )}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm">
                      {entry.org_id ? (
                        <Link to={`/orgs/${entry.org_id}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-800 dark:hover:text-primary-200">
                          {entry.org_name || truncateId(entry.org_id)}
                        </Link>
                      ) : '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400 max-w-xs truncate" title={entry.description || ''}>
                      {entry.description || '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">{formatDate(entry.created_at)}</td>
                  </tr>
                )
              })}
              {(!auditData?.audit_logs || auditData.audit_logs.length === 0) && (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">No audit logs</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {auditData && (
          <Pagination page={auditPage} totalPages={auditData.total_pages} onPageChange={setAuditPage} />
        )}
      </div>
    </div>
  )
}
