import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getAccountDetail, getAccountInstalls, getAccountAuditLogs } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-gray-900">{account.email}</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{account.id}</p>
        <div className="mt-2 flex items-center gap-2">
          <Badge>{account.account_type}</Badge>
          <span className="text-sm text-gray-500">Created {formatDate(account.created_at)}</span>
        </div>
      </div>

      {/* Roles */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Roles</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Org</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(account.roles || []).map((role) => (
                <tr key={role.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge>{role.role_type}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/orgs/${role.org_id}`} className="text-primary-600 hover:text-primary-800">
                      {role.org?.name || truncateId(role.org_id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(role.created_at)}</td>
                </tr>
              ))}
              {(!account.roles || account.roles.length === 0) && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500">No roles</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Apps */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Apps</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {apps.map((app) => (
                <tr key={app.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-medium text-gray-900">{app.name}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(app.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(app.created_at)}</td>
                </tr>
              ))}
              {apps.length === 0 && (
                <tr>
                  <td colSpan={3} className="px-4 py-8 text-center text-sm text-gray-500">No apps</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Installs */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Installs</h2>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(installsData?.installs || accountInstalls).map((install) => (
                <tr key={install.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Link to={`/installs/${install.id}`} className="text-primary-600 hover:text-primary-800 font-medium">
                      {install.name || truncateId(install.id)}
                    </Link>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(install.id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm">
                    <Badge variant="status" status={install.status}>{install.status}</Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(install.created_at)}</td>
                </tr>
              ))}
              {(installsData?.installs || accountInstalls).length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No installs</td>
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
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Audit Logs</h2>
        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
          <input
            type="text"
            value={auditEntityType}
            onChange={(e) => { setAuditEntityType(e.target.value); setAuditPage(1) }}
            placeholder="Entity type filter..."
            className="block w-48 rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <input
            type="date"
            value={startDate}
            onChange={(e) => { setStartDate(e.target.value); setAuditPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
          <input
            type="date"
            value={endDate}
            onChange={(e) => { setEndDate(e.target.value); setAuditPage(1) }}
            className="block rounded-md border-0 py-1 px-2 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
          />
        </div>
        <div className="mt-2 overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entity Type</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entity ID</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200 bg-white">
              {(auditData?.entries || []).map((entry) => (
                <tr key={entry.id}>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-900">{entry.action}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{entry.entity_type}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500 font-mono">{truncateId(entry.entity_id)}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-500">{formatDate(entry.created_at)}</td>
                </tr>
              ))}
              {(!auditData?.entries || auditData.entries.length === 0) && (
                <tr>
                  <td colSpan={4} className="px-4 py-8 text-center text-sm text-gray-500">No audit logs</td>
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
