import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router'
import { getOrgDetail, getOrgStatus, getOrgInstalls, addOrgLabels, removeOrgLabel, addSupportUsers, migrateOrgQueues } from '@/lib/admin-api'
import { Badge } from '@/components/common/Badge'
import { Pagination } from '@/components/common/Pagination'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { formatDate, truncateId } from '@/utils/format'

export const OrgDetail = () => {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const [installsPage, setInstallsPage] = useState(1)
  const [labelKey, setLabelKey] = useState('')
  const [labelValue, setLabelValue] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['org', id],
    queryFn: () => getOrgDetail(id!),
    enabled: !!id,
  })

  const { data: statusData } = useQuery({
    queryKey: ['org-status', id],
    queryFn: () => getOrgStatus(id!),
    enabled: !!id,
    refetchInterval: 20000,
  })

  const { data: installsData } = useQuery({
    queryKey: ['org-installs', id, installsPage],
    queryFn: () => getOrgInstalls(id!, { page: installsPage }),
    enabled: !!id,
  })

  const addLabelMutation = useMutation({
    mutationFn: (labels: Record<string, string>) => addOrgLabels(id!, labels),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org', id] })
      setLabelKey('')
      setLabelValue('')
    },
  })

  const removeLabelMutation = useMutation({
    mutationFn: (key: string) => removeOrgLabel(id!, key),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const supportMutation = useMutation({
    mutationFn: () => addSupportUsers(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  const migrateMutation = useMutation({
    mutationFn: () => migrateOrgQueues(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['org', id] }),
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load organization'} />
  if (!data) return null

  const { org, support_users = [] } = data
  const installs = installsData?.installs || []
  const orgLabels = org.labels || {}

  const handleAddLabel = () => {
    if (!labelKey.trim()) return
    addLabelMutation.mutate({ [labelKey.trim()]: labelValue.trim() })
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="page-heading">{org.name}</h1>
        <p className="mt-1 text-sm text-gray-500 font-mono">{org.id}</p>
        <div className="mt-2 flex items-center gap-2">
          {statusData && (
            <Badge variant="status" status={statusData.status}>{statusData.status}</Badge>
          )}
          {statusData?.status_description && (
            <span className="text-xs text-gray-500">{statusData.status_description}</span>
          )}
        </div>
      </div>

      {/* Tags (read-only) */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Tags</h2>
        <div className="mt-2 flex flex-wrap gap-2">
          {(org.tags || []).map((tag) => (
            <Badge key={tag}>{tag}</Badge>
          ))}
          {(org.tags || []).length === 0 && <span className="text-sm text-gray-500">No tags</span>}
        </div>
      </div>

      {/* Labels (editable) */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Labels</h2>
        {Object.keys(orgLabels).length > 0 ? (
          <div className="mt-2 flex flex-wrap gap-2">
            {Object.entries(orgLabels).map(([key, value]) => (
              <span key={key} className="inline-flex items-center gap-1 rounded-md bg-blue-50 border border-blue-200 px-2 py-0.5 text-xs font-mono">
                <span className="text-blue-700">{key}</span>
                <span className="text-blue-400">=</span>
                <span className="text-blue-600">{String(value)}</span>
                <button
                  onClick={() => removeLabelMutation.mutate(key)}
                  disabled={removeLabelMutation.isPending}
                  className="ml-1 text-blue-400 hover:text-red-500"
                >
                  &times;
                </button>
              </span>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No labels</p>
        )}
        <div className="mt-2 flex gap-2">
          <input
            type="text"
            value={labelKey}
            onChange={(e) => setLabelKey(e.target.value)}
            placeholder="Key"
            className="block w-32 rounded-md border-0 py-1.5 px-2.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400"
          />
          <input
            type="text"
            value={labelValue}
            onChange={(e) => setLabelValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddLabel()}
            placeholder="Value"
            className="block w-40 rounded-md border-0 py-1.5 px-2.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400"
          />
          <button
            onClick={handleAddLabel}
            disabled={addLabelMutation.isPending}
            className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            Add
          </button>
        </div>
      </div>

      {/* Support Users */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Support users</h2>
        {support_users.length > 0 ? (
          <ul className="mt-2 space-y-1">
            {support_users.map((email) => (
              <li key={email} className="text-sm text-gray-700">{email}</li>
            ))}
          </ul>
        ) : (
          <p className="mt-2 text-sm text-gray-500">No support users</p>
        )}
        <div className="mt-3">
          <button
            onClick={() => supportMutation.mutate()}
            disabled={supportMutation.isPending}
            className="rounded-md bg-primary-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            {supportMutation.isPending ? 'Adding...' : 'Add support users'}
          </button>
          {supportMutation.isSuccess && <span className="ml-2 text-sm text-green-600">Done</span>}
          {supportMutation.isError && <span className="ml-2 text-sm text-red-600">Failed</span>}
        </div>
      </div>

      {/* Actions */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Actions</h2>
        <div className="mt-2">
          <button
            onClick={() => migrateMutation.mutate()}
            disabled={migrateMutation.isPending}
            className="rounded-md bg-orange-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
          >
            {migrateMutation.isPending ? 'Migrating...' : 'Migrate queues'}
          </button>
          {migrateMutation.isSuccess && <span className="ml-2 text-sm text-green-600">Migration started</span>}
          {migrateMutation.isError && <span className="ml-2 text-sm text-red-600">Failed</span>}
        </div>
      </div>

      {/* Installs */}
      <div className="rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-semibold text-gray-900">Installs</h2>
        <div className="mt-2 table-card">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>ID</th>
                <th>Status</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {installs.map((install: any) => (
                <tr key={install.id}>
                  <td>
                    <Link to={`/installs/${install.id}`} className="text-primary-600 hover:text-primary-700 font-medium">
                      {install.name || truncateId(install.id)}
                    </Link>
                  </td>
                  <td className="text-gray-500 font-mono text-xs">{truncateId(install.id)}</td>
                  <td><Badge variant="status" status={install.status}>{install.status}</Badge></td>
                  <td className="text-gray-500">{formatDate(install.created_at)}</td>
                </tr>
              ))}
              {installs.length === 0 && (
                <tr><td colSpan={4} className="text-center text-gray-500 py-6">No installs found</td></tr>
              )}
            </tbody>
          </table>
        </div>
        {installsData && (
          <Pagination page={installsPage} totalPages={installsData.total_pages} onPageChange={setInstallsPage} />
        )}
      </div>
    </div>
  )
}
